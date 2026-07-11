# Hot Policy Reload and CLI Resilience Design

## Summary

Make manual and agent-driven Telegram flows recover cleanly from local policy
changes, negative Telegram group IDs, command vocabulary mismatches, and reply
wait timeouts. The IPC server will pick up `policy.json` changes without a
restart, peer-taking commands will accept a negative positional ID, `bot step`
will accept `--text` as an alias for `--send`, and a timeout after a successful
action will report a typed partial outcome with usable Run ID and Trace ID
metadata.

The existing immutable `policy.Enforcer` remains responsible for evaluating one
normalized policy snapshot. A new reloadable wrapper owns file-change detection
and atomically replaces the active snapshot only after a complete policy file
has been loaded and normalized successfully.

## Evidence and Current Behavior

The server currently calls `policy.LoadDefault()` once from `createIPCServer`
and stores the resulting `Enforcer` for its lifetime. Recent redacted audit and
server logs show a request for group `-5424738551` being denied by the old allow
list, followed by successful requests after an IPC-server restart.

Cobra interprets a positional token such as `-5424738551` as flags before the
command receives its arguments. The same value works as `--to=-5424738551`.

`bot step` exposes `--send`, but related send commands use text-oriented naming.
There is no `--text` alias or command-specific recovery hint today.

Reply waiting is implemented as a client-side polling phase after the Telegram
action succeeds. `WaitForReply` returns a plain error, and callers pass that
error to `Runner.Fatal`, which turns every agent-mode fatal error into
`VALIDATION`. The successful action is audited, but the overall wait timeout is
not recorded as a typed partial outcome.

## Goals

- Apply valid local policy changes to the next policy check without restarting
  the IPC server.
- Never replace the last valid policy with a malformed or unreadable update.
- Accept negative decimal Telegram peer IDs wherever the first positional
  argument is explicitly defined as a peer.
- Keep `--send` canonical for `bot step` while accepting `--text` as an alias.
- Return `TIMEOUT`, not `VALIDATION`, when reply polling reaches its deadline.
- Preserve proof that the preceding Telegram action succeeded.
- Make Run ID, Trace ID, and recovery commands available in the timeout result.
- Keep socket and HTTP policy behavior consistent.

## Non-Goals

- Do not add a general-purpose configuration watcher framework.
- Do not reload Telegram sessions or unrelated configuration.
- Do not loosen the active policy because a replacement file cannot be read or
  parsed.
- Do not redesign every CLI flag name or add aliases outside the observed
  `bot step --text` mismatch.
- Do not treat a successful action followed by a reply timeout as safe to retry
  automatically; sending or pressing again may duplicate an action.
- Do not change ordinary RPC deadline errors beyond preserving their existing
  `TIMEOUT` classification.
- Do not include message text or policy peer lists in new logs.

## Considered Policy Reload Approaches

### Load the policy on every request

This is simple and always current, but performs avoidable filesystem and JSON
work for unchanged files and repeatedly emits the same failure for a broken
update.

### Watch the policy file in a background goroutine

A filesystem watcher can react immediately, but requires another dependency or
platform-specific behavior and must handle editor-specific write, rename,
delete, and recreation sequences. It is more machinery than this small local
configuration file requires.

### Check policy content identity before each policy evaluation

This is the selected approach. A reloadable checker reads the small policy file
and obtains a content fingerprint before each `Check`. Unchanged bytes skip JSON
decoding and snapshot construction. When the content fingerprint changes, one
caller validates the captured bytes while concurrent callers continue safely.
A successful load atomically replaces the current immutable `Enforcer`.

This makes the next policy-protected request observe a completed file change,
does not need a background lifecycle, and keeps the reload boundary inside the
policy package.

## Policy Reload Architecture

Add a `policy.ReloadingEnforcer` that implements `ipc.PolicyChecker`. It owns:

- the selected policy path;
- the `PeerResolver` used to build immutable `Enforcer` snapshots;
- the last successfully observed file fingerprint;
- the current valid `Enforcer`;
- synchronization for reload checks and snapshot replacement;
- a rate-limited or state-transition-based warning path for failed reloads.

The fingerprint must distinguish a stable unchanged file from replacement,
content update, deletion, and recreation. It includes existence and a digest of
the complete bytes used for decoding. Correctness does not depend on timestamp
resolution, so a same-size, same-timestamp update cannot remain stale.

`Check` follows this flow:

1. Inspect the configured file and obtain a stable content snapshot.
2. If its digest is unchanged, use the current policy snapshot.
3. If it changed, serialize reload work and confirm the current content digest.
4. Decode, normalize, and validate those exact captured bytes.
5. Atomically publish a new immutable `Enforcer` and fingerprint.
6. Evaluate the current request against the published snapshot.

If the file does not exist, `policy.Load` continues to define the default policy
as the valid result. A confirmed deletion therefore switches to defaults. A
permission error, partial write, malformed JSON, unsupported policy version, or
other invalid content leaves the last valid snapshot active.

Startup follows the same fail-closed rule as reload. If no prior valid snapshot
exists and loading fails, the checker uses the existing conservative default
policy and logs the failure. A later corrected file is eligible for reload
without restarting the server.

Both `SocketServer` and `HTTPServer` receive the same reloadable checker from the
server construction path. The underlying IPC server interfaces do not need to
know whether a checker is static or reloadable.

## Policy Validation and Logging

Policy loading must reject structurally invalid configuration before
publication. Validation includes the supported policy version and any existing
normalization invariants. Unknown JSON fields retain the project's current
compatibility behavior unless validation already rejects them.

Log state transitions rather than every request:

- `policy reloaded` with path, version, and a non-sensitive fingerprint;
- `policy reload failed` with path and error, emitted once for a failed file
  identity rather than on every check;
- `policy reload recovered` or the next successful `policy reloaded` event when
  a previously invalid file becomes valid.

Do not log `allowPeers`, `denyPeers`, message text, or other policy contents.

## Negative Positional Peer IDs

Normalize argv before Cobra's flag parser only for commands explicitly marked
as accepting a peer in their first positional slot. A shared command annotation
or registration helper identifies that capability. When the first peer token
matches a strict negative decimal Telegram ID, rewrite it to the command's
existing `--to=<value>` form before normal flag parsing.

The normalization must:

- accept values such as `-5424738551`;
- preserve global and command flags before or after the peer;
- leave `--to=-5424738551` unchanged;
- respect `--` and ordinary negative numeric arguments that are not peer slots;
- avoid interpreting strings such as `-abc`, `-1.5`, or a negative message ID as
  peers;
- cover every existing command whose syntax makes the first positional token a
  recipient, including bot, message, chat-keyboard, send, and game helpers;
- be driven by command metadata so new peer-taking commands can opt in without
  changing a central command-name switch.

If a negative-looking token cannot be normalized for the selected command, the
Cobra error path should include the safe form `--to=-5424738551` when relevant.

## `bot step` Text Alias and Flag Guidance

Keep `--send` as the documented canonical flag and add `--text` as an accepted
alias. Internally resolve both flags into one send-text value before executing
the command.

- Only `--send`: use its value.
- Only `--text`: use its value.
- Both with identical values: accept the command.
- Both with different values: return `VALIDATION` and show a canonical
  `bot step <peer> --send <text>` example.

Unknown-flag errors should offer the nearest valid flag when the match is
unambiguous and include one command-specific example. In particular,
`bot step --txt` should point to `--text` or canonical `--send`. Suggestions
must not execute a corrected command automatically.

Generated command help, manifest/LLM guidance, README examples, and the bundled
`agent-telegram` skill continue to show `--send` while noting that `--text` is
accepted.

## Reply Timeout Contract

A reply wait is a second phase after an optional action. Reaching its deadline
is a typed timeout, not parameter validation. Introduce a typed CLI failure path
that accepts an `ipc.ErrorObject`, optional partial result, and recovery actions
instead of routing runtime failures through `Runner.Fatal`.

For a successful send or button press followed by a timeout, agent-mode output
has this shape:

```json
{
  "ok": false,
  "runId": "run_...",
  "traceId": "trace_...",
  "command": "agent-telegram bot step",
  "method": "send_message",
  "safety": "write",
  "error": {
    "code": -32004,
    "type": "TIMEOUT",
    "message": "no reply within 20s",
    "retryable": true,
    "data": {
      "actionSucceeded": true
    }
  },
  "partialResult": {
    "action": {"id": 123},
    "wait": {
      "afterMessageId": 123,
      "polls": 30,
      "timeout": "20s",
      "completed": false
    }
  },
  "diagnosis": {
    "category": "timeout",
    "summary": "The Telegram action succeeded, but no matching reply arrived before the deadline.",
    "retry": "Continue waiting or inspect the trace; do not repeat the action automatically."
  },
  "nextActions": []
}
```

The process exits nonzero because the requested compound workflow did not
complete. `actionSucceeded: true` is present only when the CLI has a successful
action result. A standalone `msg wait` timeout remains `TIMEOUT` but does not
claim an action succeeded and does not contain a fabricated action result.

The first recovery action for a partial timeout is a `msg wait` command using
the same peer, `afterMessageId`, and Run ID. A second action inspects the Trace
ID. Recovery commands must shell-quote peer values and retain the current Run ID
so follow-up activity remains correlated.

This shared behavior applies to `bot step`, `bot press`, root/send
`--wait-reply`, inline-button waits, reply-keyboard waits, and any other caller
of the common wait helper. Ordinary server-side or transport timeouts continue
through the existing RPC error path and do not include `actionSucceeded` unless
an action result is actually available.

## Observability

The successful Telegram action remains an `ok` audit event with its Run ID and
Trace ID. The compound command then records a second audit event with:

- `status: "partial"` when an action succeeded but waiting timed out;
- `errorType: "TIMEOUT"` and timeout error code;
- the same Run ID and Trace ID;
- method/safety metadata for the action;
- a redacted result summary containing the action ID and wait metadata.

A standalone wait timeout uses `status: "error"`, because there is no partial
action success. CLI logs include the wait phase, `afterMessageId`, poll count,
timeout, Run ID, and Trace ID. New events never include message text.

Run ID and Trace ID remain top-level fields in every agent-mode error envelope,
including local CLI failures. Timeout recovery actions must never contain empty
identifier placeholders; omit an unavailable identifier or generate it through
the existing runner initialization path.

## Components

### Reloadable policy checker

Keep file observation and snapshot publication in `internal/policy`. Preserve
`Enforcer` as a small immutable evaluator that can still be unit-tested without
filesystem state.

### Peer argv normalization

Keep command annotations and argv transformation in a focused CLI utility. The
root execution path invokes it before Cobra parses flags. Individual commands
only declare that their first positional argument is a peer.

### Typed local failure output

Extend `Runner` with a typed failure method rather than broadening `Fatal`,
whose existing callers represent validation failures. The typed path owns the
error envelope, optional `partialResult`, audit status, diagnosis, and recovery
actions.

### Wait outcome

Represent reply polling results with enough structured metadata to distinguish
success from deadline expiry. Callers should not parse error strings to decide
whether a timeout occurred.

## Testing

### Policy unit tests

- unchanged files reuse the current snapshot;
- adding and removing an allowed peer changes the next decision;
- peer-type changes take effect without restart;
- malformed and partially written JSON preserve the last valid snapshot;
- a corrected file recovers automatically;
- deletion activates defaults and recreation activates the new file;
- same-size rapid replacements are detected;
- unsupported versions are rejected;
- concurrent `Check` calls never observe a partially constructed policy;
- repeated checks of one broken identity do not flood logs or reload work.

### IPC integration tests

Start an IPC server with a temporary policy file, observe a denied request,
rewrite the allow list while the server remains running, and observe the next
request succeed. Exercise both the socket policy path and HTTP checker behavior
where their construction differs.

### CLI parsing tests

- every annotated peer-taking command accepts a negative first positional ID;
- flags work before and after that ID;
- explicit `--to=-5424738551` remains valid;
- `--` behavior is unchanged;
- negative message IDs and other numeric arguments are not rewritten;
- malformed negative-looking strings receive a targeted hint;
- `bot step --send`, `--text`, identical dual values, and conflicting dual
  values follow the defined contract;
- unknown bot-step flags produce a suggestion and canonical example.

### Timeout contract tests

- a successful send plus timeout returns `TIMEOUT`, `partialResult`, and
  `actionSucceeded: true`;
- successful inline and reply-keyboard actions use the same contract;
- standalone `msg wait` timeout omits action success;
- timeout is retryable but its diagnosis warns against repeating the action;
- recovery commands include `afterMessageId`, quoted peer, Run ID, and Trace ID;
- audit records `ok` followed by `partial` with matching identifiers;
- non-timeout validation still returns `VALIDATION`;
- ordinary RPC timeout remains `TIMEOUT` without a fabricated partial result.

## Documentation

Update README, DEVELOPMENT guidance, command help, generated LLM text, manifest
examples, and the bundled skill. Document:

- policy changes are applied automatically to the next protected request;
- invalid updates preserve the last valid policy and are visible in server
  logs;
- negative positional peer IDs and the explicit `--to=<negative-id>` form;
- canonical `bot step --send` plus the accepted `--text` alias;
- partial timeout semantics and the importance of Run ID and Trace ID;
- the safe recovery rule: continue waiting or inspect the trace before deciding
  whether to repeat a write action.

## Acceptance Criteria

- A valid policy edit changes the next protected socket and HTTP request without
  restarting the server.
- An invalid policy edit never replaces or weakens the last valid policy.
- Recovery from an invalid edit requires only correcting the file, not a
  restart.
- `agent-telegram bot step -5424738551 --send /start` reaches command execution
  with peer `-5424738551`.
- `bot step --text hello` behaves like `bot step --send hello`, and conflicting
  aliases fail with a precise example.
- A reply deadline is reported as `TIMEOUT`, never `VALIDATION`.
- When the action succeeded, the timeout output contains its partial result,
  Run ID, Trace ID, and safe follow-up commands.
- Audit and logs distinguish action success from compound-command timeout
  without exposing message text or policy lists.
- Existing positive IDs, usernames, explicit `--to`, non-peer numeric
  arguments, RPC error behavior, and policy enforcement remain compatible.
