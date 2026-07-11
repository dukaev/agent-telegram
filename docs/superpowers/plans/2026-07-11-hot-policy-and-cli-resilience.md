# Hot Policy Reload and CLI Resilience Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Apply policy edits without restarting the IPC server and make negative peer IDs, `bot step` text input, reply timeouts, and Run/Trace diagnostics reliable for manual and agent-driven flows.

**Architecture:** Keep `policy.Enforcer` immutable and wrap it in a file-backed `ReloadingEnforcer` that hashes the small policy file before each check and only publishes valid snapshots. Normalize negative peer argv before Cobra flag parsing using opt-in command metadata. Route reply deadline expiry through a typed runner failure carrying a partial action result, correlated audit metadata, and safe recovery commands.

**Tech Stack:** Go 1.25.4, Cobra 1.10.2, pflag 1.0.9, JSON-RPC over Unix sockets/HTTP, `log/slog`, the existing observability journal, and the Go standard library (`crypto/sha256`, `sync`).

## Global Constraints

- Add no new runtime dependencies.
- Keep `--send` canonical; `--text` is an accepted `bot step` alias only.
- A malformed or unreadable replacement policy never replaces the last valid in-memory policy.
- A confirmed missing policy file activates `policy.Default()`.
- A reply timeout is `TIMEOUT`, remains nonzero-exit, and is never reported as `VALIDATION`.
- Never retry a successful write action automatically after its reply wait times out.
- Preserve Run ID and Trace ID across recovery commands.
- New logs and audit summaries must not contain message text, `allowPeers`, or `denyPeers`.
- Keep existing username, positive ID, explicit `--to`, RPC timeout, and policy enforcement behavior compatible.

---

### Task 1: Reloadable policy snapshots

**Files:**
- Create: `internal/policy/reloading.go`
- Create: `internal/policy/reloading_test.go`
- Modify: `internal/policy/policy.go:65-111`
- Modify: `internal/policy/policy_test.go`

**Interfaces:**
- Produces: `func Parse(data []byte) (Policy, error)`
- Produces: `func (p Policy) Validate() error`
- Produces: `func NewReloadingEnforcer(path string, resolver PeerResolver) *ReloadingEnforcer`
- Produces: `func (e *ReloadingEnforcer) Check(context.Context, string, json.RawMessage) error`
- Preserves: `func Load(path string) (Policy, error)` and `func NewEnforcer(Policy, PeerResolver) *Enforcer`

- [ ] **Step 1: Write failing parse and version-validation tests**

Add table-driven tests to `internal/policy/policy_test.go` that prove valid JSON is normalized, version `2` is rejected, malformed JSON is rejected, and a missing file still returns `Default()`:

```go
func TestParseValidatesAndNormalizesPolicy(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr string
	}{
		{name: "valid", body: `{"version":1,"allowPeers":["Ada"]}`},
		{name: "unsupported version", body: `{"version":2}`, wantErr: "unsupported policy version 2"},
		{name: "malformed", body: `{`, wantErr: "parse policy"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse([]byte(tt.body))
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil || len(got.AllowPeers) != 1 || got.AllowPeers[0] != "@ada" {
				t.Fatalf("policy = %+v, error = %v", got, err)
			}
		})
	}
}
```

- [ ] **Step 2: Run the focused policy tests and verify failure**

Run: `go test ./internal/policy -run 'TestParseValidatesAndNormalizesPolicy' -count=1`

Expected: FAIL because `Parse` and `Policy.Validate` do not exist.

- [ ] **Step 3: Extract parsing and validation from `Load`**

Implement the production boundary in `internal/policy/policy.go`:

```go
func Parse(data []byte) (Policy, error) {
	p := Default()
	if err := json.Unmarshal(data, &p); err != nil {
		return Policy{}, fmt.Errorf("parse policy: %w", err)
	}
	p.Normalize()
	if err := p.Validate(); err != nil {
		return Policy{}, err
	}
	return p, nil
}

func (p Policy) Validate() error {
	if p.Version != version {
		return fmt.Errorf("unsupported policy version %d", p.Version)
	}
	return nil
}
```

Change `Load` to return `Default()` on `os.IsNotExist` and otherwise call `Parse(data)`. Do not change JSON unknown-field compatibility.

- [ ] **Step 4: Write failing reload, recovery, deletion, and concurrency tests**

Create `internal/policy/reloading_test.go` with a helper that writes policy JSON using mode `0600`, then cover these exact transitions:

```go
func TestReloadingEnforcerAppliesOnlyValidSnapshots(t *testing.T) {
	path := filepath.Join(t.TempDir(), "policy.json")
	writePolicy := func(body string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	params := json.RawMessage(`{"peer":"@ada","message":"hello"}`)
	writePolicy(`{"version":1,"allowPeers":["@grace"]}`)
	enforcer := NewReloadingEnforcer(path, nil)
	if err := enforcer.Check(context.Background(), "send_message", params); err == nil {
		t.Fatal("@ada should initially be denied")
	}

	writePolicy(`{"version":1,"allowPeers":["@ada"]}`)
	if err := enforcer.Check(context.Background(), "send_message", params); err != nil {
		t.Fatalf("valid edit was not applied: %v", err)
	}

	writePolicy(`{"version":1,"allowPeers":[`)
	if err := enforcer.Check(context.Background(), "send_message", params); err != nil {
		t.Fatalf("malformed edit replaced last valid snapshot: %v", err)
	}

	writePolicy(`{"version":1,"allowPeers":["@grace"]}`)
	if err := enforcer.Check(context.Background(), "send_message", params); err == nil {
		t.Fatal("corrected edit was not applied")
	}

	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := enforcer.Check(context.Background(), "send_message", params); err != nil {
		t.Fatalf("deletion should activate defaults: %v", err)
	}
}
```

Add separate tests for same-size rapid replacement, unsupported version recovery, 32 concurrent `Check` calls during rewrites, and one warning per failed digest using a temporary default `slog` handler.

- [ ] **Step 5: Run reload tests and verify failure**

Run: `go test ./internal/policy -run 'TestReloading|TestParse' -race -count=1`

Expected: FAIL because `NewReloadingEnforcer` is undefined.

- [ ] **Step 6: Implement the reloadable checker**

Create `internal/policy/reloading.go` around these concrete state types:

```go
type fileFingerprint struct {
	exists bool
	digest [sha256.Size]byte
}

type ReloadingEnforcer struct {
	path        string
	resolver    PeerResolver
	mu          sync.Mutex
	current     *Enforcer
	attempted   fileFingerprint
	hasAttempt  bool
	lastFailure string
}

func NewReloadingEnforcer(path string, resolver PeerResolver) *ReloadingEnforcer {
	e := &ReloadingEnforcer{
		path:     path,
		resolver: resolver,
		current:  NewEnforcer(Default(), resolver),
	}
	e.refresh()
	return e
}

func (e *ReloadingEnforcer) Check(ctx context.Context, method string, params json.RawMessage) error {
	e.refresh()
	e.mu.Lock()
	current := e.current
	e.mu.Unlock()
	return current.Check(ctx, method, params)
}
```

`refresh` must read the complete file, map `os.IsNotExist` to `(Default(), fileFingerprint{exists:false})`, hash existing bytes before parsing, suppress repeated work for the same attempted fingerprint, leave `current` unchanged on error, and log only digest state transitions. Never log policy contents.

- [ ] **Step 7: Run policy tests with the race detector**

Run: `go test ./internal/policy -race -count=1`

Expected: PASS.

- [ ] **Step 8: Commit the policy unit**

```bash
git add internal/policy/policy.go internal/policy/policy_test.go internal/policy/reloading.go internal/policy/reloading_test.go
git commit -m "feat: hot reload local policy snapshots"
```

---

### Task 2: Wire hot reload through socket and HTTP policy paths

**Files:**
- Modify: `cmd/serve.go:373-505`
- Modify: `cmd/serve_api.go`
- Create: `internal/ipc/policy_reload_integration_test.go`

**Interfaces:**
- Consumes: `policy.NewReloadingEnforcer(path, resolver)` from Task 1
- Preserves: `ipc.PolicyChecker`, `SocketServer.SetPolicyChecker`, and `HTTPServer.SetPolicyChecker`
- Produces: one live checker instance per running server surface

- [ ] **Step 1: Write the socket integration test and a failing production-wiring assertion**

Create an external-package test (`package ipc_test`) that starts a real temporary Unix socket, registers `send_message`, and changes the policy while the same server stays alive:

```go
func TestSocketServerUsesReloadedPolicyWithoutRestart(t *testing.T) {
	dir := t.TempDir()
	policyPath := filepath.Join(dir, "policy.json")
	socketPath := filepath.Join(dir, "agent-telegram.sock")
	if err := os.WriteFile(policyPath, []byte(`{"version":1,"allowPeers":["@grace"]}`), 0o600); err != nil {
		t.Fatal(err)
	}

	srv := ipc.NewSocketServer(socketPath)
	srv.SetPolicyChecker(policy.NewReloadingEnforcer(policyPath, nil))
	srv.Register("send_message", func(context.Context, json.RawMessage) (any, *ipc.ErrorObject) {
		return map[string]any{"id": 1}, nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = srv.Start(ctx) }()
	waitForSocket(t, socketPath)

	client := ipc.NewClient(socketPath)
	params := map[string]any{"peer": "@ada", "message": "hello"}
	if _, rpcErr := client.Call("send_message", params); rpcErr == nil || rpcErr.Code != ipc.ErrCodePolicyDenied {
		t.Fatalf("initial error = %+v, want policy denied", rpcErr)
	}
	if err := os.WriteFile(policyPath, []byte(`{"version":1,"allowPeers":["@ada"]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, rpcErr := client.Call("send_message", params); rpcErr != nil {
		t.Fatalf("request after edit failed without restart: %+v", rpcErr)
	}
}
```

`waitForSocket` polls `os.Stat` for at most two seconds with a 10 ms interval and fails with the last error; it must not use a fixed sleep.

Also add this assertion to `cmd/serve_test.go`:

```go
func TestLoadPolicyCheckerReturnsReloadingEnforcer(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	checker := loadPolicyChecker(nil)
	if _, ok := checker.(*policy.ReloadingEnforcer); !ok {
		t.Fatalf("checker = %T, want *policy.ReloadingEnforcer", checker)
	}
}
```

- [ ] **Step 2: Run the integration and wiring tests and verify the wiring failure**

Run: `go test ./internal/ipc ./cmd -run 'TestSocketServerUsesReloadedPolicyWithoutRestart|TestLoadPolicyCheckerReturnsReloadingEnforcer' -count=1`

Expected: the socket integration test passes and the command test fails because `loadPolicyChecker` still returns `*policy.Enforcer`.

- [ ] **Step 3: Generalize the production checker constructor**

Change `loadPolicyChecker` from a concrete `*telegram.Client` parameter to `policy.PeerResolver`. Resolve `policy.DefaultPath()` inside this helper exactly once, return a static default `Enforcer` only when path resolution itself fails, and otherwise return `policy.NewReloadingEnforcer(path, resolver)`.

- [ ] **Step 4: Confirm both server surfaces use the shared helper**

Change the construction boundary to this behavior:

```go
func loadPolicyChecker(tgClient policy.PeerResolver) ipc.PolicyChecker {
	path, err := policy.DefaultPath()
	if err != nil {
		slog.Warn("failed to resolve local policy path, using defaults", "error", err)
		return policy.NewEnforcer(policy.Default(), tgClient)
	}
	return policy.NewReloadingEnforcer(path, tgClient)
}
```

Use the same helper for the Unix socket server and `serve-api` HTTP server. Do not add reload behavior to `ipc.Server`; it already depends on the `PolicyChecker` interface.

- [ ] **Step 5: Run socket, HTTP, and command server tests**

Run: `go test ./internal/ipc ./cmd -run 'Policy|Server|HTTP' -race -count=1`

Expected: PASS, including the no-restart request transition.

- [ ] **Step 6: Commit server wiring**

```bash
git add cmd/serve.go cmd/serve_api.go cmd/serve_test.go internal/ipc/policy_reload_integration_test.go
git commit -m "feat: apply policy edits without server restart"
```

---

### Task 3: Accept negative positional peer IDs before Cobra flag parsing

**Files:**
- Create: `internal/cliutil/peer_args.go`
- Create: `internal/cliutil/peer_args_test.go`
- Modify: `cmd/root.go:194-220`
- Modify: `cmd/bot/bot.go`
- Modify: `cmd/send/send.go`
- Modify: `cmd/send/text.go`
- Modify: `cmd/send/poll.go`
- Modify: `cmd/send/contact.go`
- Modify: `cmd/send/location.go`
- Modify: `cmd/send/dice.go`
- Modify: `cmd/game/dice.go`
- Modify: `cmd/chat/keyboard.go`
- Modify: `cmd/message/list.go`
- Modify: `cmd/message/replies.go`
- Modify: `cmd/message/keyboard.go`
- Modify: `cmd/message/press_keyboard.go`
- Modify: `cmd/message/wait.go`
- Modify: `cmd/message/reply_comment.go`

**Interfaces:**
- Produces: `func MarkFirstArgPeer(cmd *cobra.Command)`
- Produces: `func AcceptsFirstArgPeer(cmd *cobra.Command) bool`
- Produces: `func NormalizeNegativePeerArgs(root *cobra.Command, args []string) []string`
- Consumes: Cobra command annotations and inherited/local pflag metadata
- Guarantee: only the first positional peer of an opted-in command is rewritten to `--to=<negative-decimal>`

- [ ] **Step 1: Write failing argv-normalization tests**

Create an isolated Cobra tree in `internal/cliutil/peer_args_test.go` and cover the exact matrix:

```go
func TestNormalizeNegativePeerArgs(t *testing.T) {
	root := &cobra.Command{Use: "agent-telegram"}
	root.PersistentFlags().String("run-id", "", "")
	step := &cobra.Command{Use: "step [peer]"}
	step.Flags().String("send", "", "")
	step.Flags().String("to", "", "")
	MarkFirstArgPeer(step)
	bot := &cobra.Command{Use: "bot"}
	bot.AddCommand(step)
	root.AddCommand(bot)

	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{name: "plain", in: []string{"bot", "step", "-5424738551", "--send", "/start"}, want: []string{"bot", "step", "--to=-5424738551", "--send", "/start"}},
		{name: "global flag", in: []string{"--run-id", "run-1", "bot", "step", "-5424738551"}, want: []string{"--run-id", "run-1", "bot", "step", "--to=-5424738551"}},
		{name: "explicit to", in: []string{"bot", "step", "--to=-5424738551"}, want: []string{"bot", "step", "--to=-5424738551"}},
		{name: "separator", in: []string{"bot", "step", "--", "-5424738551"}, want: []string{"bot", "step", "--", "-5424738551"}},
		{name: "not decimal", in: []string{"bot", "step", "-abc"}, want: []string{"bot", "step", "-abc"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeNegativePeerArgs(root, tt.in); !slices.Equal(got, tt.want) {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
```

Add cases proving `--after-id -5` is skipped as a flag value, negative message/button indices on unannotated commands are untouched, flags can follow the rewritten peer, and the input slice is not mutated.

- [ ] **Step 2: Run parser tests and verify failure**

Run: `go test ./internal/cliutil -run TestNormalizeNegativePeerArgs -count=1`

Expected: FAIL because the annotation and normalizer do not exist.

- [ ] **Step 3: Implement opt-in metadata and flag-aware scanning**

Use one annotation key and a strict decimal matcher:

```go
const firstArgPeerAnnotation = "agent-telegram.io/first-arg-peer"

var negativePeerPattern = regexp.MustCompile(`^-[0-9]+$`)

func MarkFirstArgPeer(cmd *cobra.Command) {
	if cmd.Annotations == nil {
		cmd.Annotations = map[string]string{}
	}
	cmd.Annotations[firstArgPeerAnnotation] = "true"
}
```

`NormalizeNegativePeerArgs` must copy the input, locate the final command through the Cobra tree, scan only that command's argument tail, skip known long/short flag values using local plus inherited flag definitions, stop at `--`, and rewrite only the first positional token when it matches `negativePeerPattern`. It must never parse or mutate flag state.

- [ ] **Step 4: Mark every current first-positional-peer command**

Call `cliutil.MarkFirstArgPeer` during command registration for these command paths:

```text
agent-telegram bot step
agent-telegram bot press
agent-telegram send
agent-telegram send text
agent-telegram send poll
agent-telegram send contact
agent-telegram send location
agent-telegram send dice
agent-telegram game dice
agent-telegram chat keyboard
agent-telegram msg list
agent-telegram msg replies
agent-telegram msg inspect-keyboard
agent-telegram msg press-keyboard
agent-telegram msg wait
agent-telegram msg reply-comment
```

Do not annotate commands whose positional number is a message ID and whose peer is flag-only, such as `msg get`, `msg press-button`, and `send update`.

- [ ] **Step 5: Invoke normalization before schema detection and `Execute`**

At the top of `cmd.Execute`, compute one normalized argv slice and use it for both the existing `--schema` preflight and Cobra execution:

```go
func Execute() {
	args := cliutil.NormalizeNegativePeerArgs(RootCmd, os.Args[1:])
	RootCmd.SetArgs(args)
	if hasFlag(args, "--schema") {
		cmd, _, _ := RootCmd.Find(args)
		if cmd != nil && cmd != RootCmd {
			cliutil.PrintCommandSchema(cmd)
		}
	}
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
```

- [ ] **Step 6: Add a command-registry contract test**

In `cmd/registry_contract_test.go`, assert that all 16 paths above carry the annotation through a new exported `cliutil.AcceptsFirstArgPeer(cmd)` query, and that `msg get`, `msg press-button`, and `send update` do not.

- [ ] **Step 7: Run parser and command surface tests**

Run: `go test ./internal/cliutil ./cmd/... -run 'NegativePeer|FirstArgPeer|ExpectedSurface' -count=1`

Expected: PASS.

- [ ] **Step 8: Commit negative peer parsing**

```bash
git add internal/cliutil/peer_args.go internal/cliutil/peer_args_test.go cmd/root.go cmd/registry_contract_test.go cmd/bot cmd/send cmd/game/dice.go cmd/chat/keyboard.go cmd/message
git commit -m "fix: accept negative positional Telegram peer IDs"
```

---

### Task 4: Add `bot step --text` and actionable flag errors

**Files:**
- Create: `internal/cliutil/flag_errors.go`
- Create: `internal/cliutil/flag_errors_test.go`
- Modify: `cmd/root.go`
- Modify: `cmd/bot/bot.go:13-100`
- Modify: `cmd/bot/bot_test.go`

**Interfaces:**
- Produces: `func FlagErrorWithHints(cmd *cobra.Command, err error) error`
- Produces: pure `func resolveStepText(cmd *cobra.Command, sendValue, textValue string) (string, error)`
- Preserves: canonical help/examples use `--send`

- [ ] **Step 1: Write failing bot alias resolution tests**

Add this table to `cmd/bot/bot_test.go`:

```go
func TestResolveStepText(t *testing.T) {
	tests := []struct {
		name, send, text, want, wantErr string
		sendChanged, textChanged       bool
	}{
		{name: "send", send: "hello", want: "hello", sendChanged: true},
		{name: "text", text: "hello", want: "hello", textChanged: true},
		{name: "same", send: "hello", text: "hello", want: "hello", sendChanged: true, textChanged: true},
		{name: "conflict", send: "a", text: "b", sendChanged: true, textChanged: true, wantErr: "use only --send or --text"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "step"}
			cmd.Flags().String("send", "", "")
			cmd.Flags().String("text", "", "")
			if tt.sendChanged { _ = cmd.Flags().Set("send", tt.send) }
			if tt.textChanged { _ = cmd.Flags().Set("text", tt.text) }
			got, err := resolveStepText(cmd, tt.send, tt.text)
			if got != tt.want || (tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr))) {
				t.Fatalf("got %q, error %v", got, err)
			}
		})
	}
}
```

- [ ] **Step 2: Write failing unknown-flag hint tests**

In `internal/cliutil/flag_errors_test.go`, execute an isolated `bot step --txt hello`, assert the error retains `unknown flag: --txt`, suggests `--text` or `--send`, and includes `agent-telegram bot step <peer> --send <text>`. Add a negative-ID-shaped error case that includes `--to=-5424738551`. Assert unrelated errors are returned unchanged.

- [ ] **Step 3: Run the focused tests and verify failure**

Run: `go test ./cmd/bot ./internal/cliutil -run 'ResolveStepText|FlagErrorWithHints' -count=1`

Expected: FAIL because the alias resolver and hint formatter do not exist.

- [ ] **Step 4: Implement the alias without duplicating command behavior**

Add `stepText string`, register:

```go
StepCmd.Flags().StringVar(&stepText, "text", "", "Alias for --send")
```

At the start of `runStep`, create the runner, call `resolveStepText`, and route a conflict through `runner.Fatal` with this exact message:

```text
use only --send or --text when values differ; example: agent-telegram bot step <peer> --send <text>
```

Use the resolved string for `send_message`; keep `--send` in `nextActions` and examples.

- [ ] **Step 5: Implement nearest-flag and command-example guidance**

`FlagErrorWithHints` should parse only Cobra/pflag unknown-flag errors, calculate Levenshtein distance against local and inherited long flag names, add a suggestion only when the nearest match is unique and distance is at most two, and append the command's `Example` or the explicit bot-step example. For a token matching `^-[0-9]+$`, append `negative peer IDs must use --to=<id> if this command does not accept a positional peer`.

Install it once with `RootCmd.SetFlagErrorFunc(cliutil.FlagErrorWithHints)`. Do not auto-correct or execute anything.

- [ ] **Step 6: Run bot and CLI error tests**

Run: `go test ./cmd/bot ./internal/cliutil ./cmd -run 'StepText|FlagError|AgenticContract' -count=1`

Expected: PASS, with `--send` still present and `--text` added only to `bot step`.

- [ ] **Step 7: Commit alias and hint behavior**

```bash
git add internal/cliutil/flag_errors.go internal/cliutil/flag_errors_test.go cmd/root.go cmd/bot/bot.go cmd/bot/bot_test.go
git commit -m "feat: guide bot step text and flag usage"
```

---

### Task 5: Report reply deadlines as correlated partial timeouts

**Files:**
- Modify: `cmd/send/wait.go`
- Create: `cmd/send/wait_test.go`
- Modify: `cmd/bot/bot.go`
- Modify: `cmd/message/wait.go`
- Modify: `cmd/message/buttons.go`
- Modify: `cmd/message/press_keyboard.go`
- Modify: `internal/cliutil/runner.go:345-420`
- Modify: `internal/cliutil/runner_errors.go`
- Modify: `internal/cliutil/runner_test.go`
- Modify: `internal/cliutil/runner_call.go`
- Modify: `internal/observability/audit_test.go`

**Interfaces:**
- Produces: `type send.WaitOutcome struct { Reply any; AfterMessageID int64; Polls int; Timeout time.Duration; Completed bool }`
- Changes: `func send.WaitForReply(poller ReplyPoller, peer string, afterMessageID int64, timeout time.Duration) WaitOutcome`
- Produces: `func send.FailReplyTimeout(runner *cliutil.Runner, peer string, action any, outcome WaitOutcome)`
- Produces: `type cliutil.FailureDetails struct { PartialResult any; NextActions []map[string]any; AuditStatus string; AuditSummary map[string]any }`
- Produces: `func (r *Runner) FailTyped(err *ipc.ErrorObject, details FailureDetails)`
- Produces: `func (r *Runner) RunID() string` and `func (r *Runner) TraceID() string`

- [ ] **Step 1: Write deterministic wait outcome tests with a fake poller**

Define the narrow poller interface in `cmd/send/wait.go`:

```go
type ReplyPoller interface {
	CallInternal(method string, params any) any
}
```

Create `cmd/send/wait_test.go` with a fake returning queued `get_messages` results. Test one incoming message after the action, outgoing messages being ignored, and a one-nanosecond deadline producing:

```go
WaitOutcome{
	AfterMessageID: 123,
	Timeout:        time.Nanosecond,
	Completed:      false,
}
```

Assert poll count is stable by allowing injection of a package-level `waitNow`/`waitSleep` test clock rather than relying on wall-clock sleeps.

- [ ] **Step 2: Write failing typed-envelope and audit tests**

Extend `internal/cliutil/runner_test.go` to call `FailTyped` with a timeout and capture stdout plus an audit journal under temporary `HOME`. Assert:

```go
if body.Error.Type != ipc.ErrorTypeTimeout || !body.Error.Retryable {
	t.Fatalf("error = %+v, want retryable TIMEOUT", body.Error)
}
if !body.Error.Data.ActionSucceeded || body.PartialResult.Wait.Completed {
	t.Fatalf("partial timeout = %+v", body)
}
if body.RunID != "run-test" || body.TraceID != "trace-test" {
	t.Fatalf("correlation IDs = %+v", body)
}
```

Read the audit file and assert the final event has `Status == "partial"`, `ErrorType == "TIMEOUT"`, and matching identifiers. Add a standalone-wait case using status `error` and no `actionSucceeded` field.

- [ ] **Step 3: Run wait and runner tests and verify failure**

Run: `go test ./cmd/send ./internal/cliutil -run 'WaitOutcome|TypedTimeout|PartialAudit' -count=1`

Expected: FAIL because `WaitOutcome`, `FailureDetails`, and `FailTyped` do not exist.

- [ ] **Step 4: Implement structured polling without string classification**

Change `WaitForReply` to return a `WaitOutcome` on both success and deadline. On success set `Completed: true` and include `Reply`; on deadline set `Completed: false`, the final poll count, exact timeout, and `AfterMessageID`. Remove `fmt.Errorf("no reply within %s", timeout)` from this layer.

- [ ] **Step 5: Implement typed local failures in Runner**

Add the public details type and make `FailTyped`:

1. write a CLI log event with Run ID, Trace ID, method, `error_type`, and redacted summary;
2. append an audit event using `AuditStatus`, `AuditSummary`, and the current action method/safety;
3. print the existing agent error envelope plus `partialResult` and override `nextActions` when supplied;
4. print a concise human error and recovery commands outside agent mode;
5. exit with code 1 through `cliutil.Exit`.

Add a `TIMEOUT` branch to `diagnosisForError`:

```go
case ipc.ErrorTypeTimeout:
	return map[string]any{
		"category": "timeout",
		"summary":  "The requested wait did not complete before the deadline.",
		"retry":    "Continue waiting or inspect the trace before repeating a write action.",
	}
```

Keep `Fatal` mapped to `VALIDATION`; only runtime timeout callers use `FailTyped`.

- [ ] **Step 6: Build safe timeout recovery actions**

In `send.FailReplyTimeout`, create `ipc.NewTypedError(ipc.ErrCodeTimeout, ipc.ErrorTypeTimeout, ...)`. Include `actionSucceeded: true` only when `action != nil`. Construct the partial result from the action and wait metadata. Use these exact action kinds:

```text
wait_for_reply   agent-telegram msg wait <quoted-peer> --after-id <id> --timeout <duration> --agent --run-id <runId>
inspect_trace    agent-telegram trace inspect <traceId> --agent --run-id <runId>
```

Move the existing shell quoting helper from `cmd/bot` to `internal/cliutil` as `ShellArg` so negative IDs, usernames, spaces, quotes, `$`, and backticks are safe in generated commands.

- [ ] **Step 7: Migrate every shared wait caller**

Update these flows to inspect `outcome.Completed` and call `FailReplyTimeout` on deadline:

```text
agent-telegram send --wait-reply
agent-telegram bot step --wait-reply
agent-telegram bot press --wait-reply
agent-telegram msg press-button --wait-reply
agent-telegram msg press-keyboard --wait-reply
agent-telegram msg wait
```

For `msg wait`, pass `action=nil`, use audit status `error`, and omit `partialResult.action`. For action-based commands, preserve the successful action object and use audit status `partial`.

- [ ] **Step 8: Add bot-level timeout contract tests**

Extend `cmd/bot/bot_test.go` and `cmd/message/message_test.go` to verify the timeout helper is reached with the action result and `afterMessageId`, while success still builds bot state and `nextActions`. Use fake pollers/runner helpers introduced in this task; do not call Telegram.

- [ ] **Step 9: Run timeout, observability, bot, message, and send tests**

Run: `go test ./cmd/send ./cmd/bot ./cmd/message ./internal/cliutil ./internal/observability -race -count=1`

Expected: PASS. Agent timeout JSON contains `TIMEOUT`, top-level Run/Trace IDs, a partial result only when an action succeeded, and both recovery commands.

- [ ] **Step 10: Commit the timeout contract**

```bash
git add cmd/send/wait.go cmd/send/wait_test.go cmd/bot/bot.go cmd/bot/bot_test.go cmd/message internal/cliutil/runner.go internal/cliutil/runner_errors.go internal/cliutil/runner_call.go internal/cliutil/runner_test.go internal/observability/audit_test.go
git commit -m "fix: report reply waits as partial timeouts"
```

---

### Task 6: Document the new contract and run full verification

**Files:**
- Modify: `README.md`
- Modify: `DEVELOPMENT.md`
- Modify: `TESTING_CHECKLIST.md`
- Modify: `internal/docs/docs.go`
- Modify: `internal/docs/docs_test.go`
- Modify: `internal/skills/bundled/agent-telegram/SKILL.md`
- Modify: generated README blocks through `go run . docs generate --target README.md`

**Interfaces:**
- Documents: hot reload, invalid-policy fallback, negative positional IDs, `--text`, partial timeout semantics, and Run/Trace recovery
- Preserves: generated documentation checks and bundled skill onboarding tests

- [ ] **Step 1: Write failing documentation contract assertions**

Extend `internal/docs/docs_test.go` so generated LLM Markdown must contain all of:

```text
policy changes are applied without restarting the server
--to=-5424738551
bot step
--send
--text
partial timeout
do not repeat the action automatically
trace inspect
```

Extend `internal/skills/onboarding_test.go` or the existing bundled-skill content test to assert the bundled skill mentions hot policy reload and timeout recovery while retaining canonical `bot step @bot --send`.

- [ ] **Step 2: Run documentation tests and verify failure**

Run: `go test ./internal/docs ./internal/skills -run 'LLM|Bundled|Onboarding' -count=1`

Expected: FAIL because the new guidance is absent.

- [ ] **Step 3: Update human and generated guidance sources**

Add concise sections to README and DEVELOPMENT covering:

```text
- policy.json is checked before each protected request;
- malformed updates keep the last valid policy and emit a server warning;
- negative peer IDs work positionally on peer-taking commands, while --to=<id> remains universal;
- bot step documents --send and accepts --text;
- reply timeouts are retryable waits, not proof that the action failed;
- use the returned msg wait and trace inspect commands before repeating a write.
```

Add the same operational rules to `internal/docs/docs.go` and the bundled skill without exposing local peer lists.

- [ ] **Step 4: Regenerate README blocks and verify they are current**

Run: `go run . docs generate --target README.md`

Expected: JSON result with `"ok": true` and a boolean `changed` field.

Run: `go run . docs check --target README.md`

Expected: JSON result with `"ok": true`.

- [ ] **Step 5: Run formatting and static checks**

Run: `git diff --name-only ce5bdbe..HEAD -- '*.go' | xargs gofmt -w`

Expected: no output.

Run: `go vet ./...`

Expected: exit 0.

- [ ] **Step 6: Run the complete test suite with race-sensitive packages repeated**

Run: `go test ./... -count=1`

Expected: PASS.

Run: `go test ./internal/policy ./internal/ipc ./internal/cliutil ./cmd/send ./cmd/bot ./cmd/message -race -count=3`

Expected: PASS on all three runs with no race report.

- [ ] **Step 7: Perform local CLI smoke checks without sending a Telegram action**

Run: `go run . bot step --help`

Expected: help shows canonical `--send`, accepted `--text`, and the negative peer example.

Run: `go run . bot step -5424738551 --send a --text b --agent --run-id smoke-negative-peer`

Expected: parsing reaches bot-step validation and returns `error.type == "VALIDATION"` for the conflicting aliases; no Telegram write occurs.

Run: `go run . bot step @example --txt hello --agent --run-id smoke-flag-hint`

Expected: nonzero exit with an unknown-flag suggestion for `--text` or `--send` and the canonical example; no Telegram write occurs.

- [ ] **Step 8: Update the manual test checklist**

Add unchecked release/manual items to `TESTING_CHECKLIST.md` for editing a temporary allow list while the server stays running, correcting malformed policy JSON, exercising a negative group ID positionally, and observing a real bot reply timeout with matching Run ID/Trace ID in `audit`, `logs`, and `trace inspect`.

- [ ] **Step 9: Commit documentation and verification updates**

```bash
git add README.md DEVELOPMENT.md TESTING_CHECKLIST.md internal/docs internal/skills/bundled/agent-telegram/SKILL.md internal/skills/onboarding_test.go
git commit -m "docs: explain hot policy and partial timeout recovery"
```

- [ ] **Step 10: Confirm the worktree contains only intentional changes**

Run: `git status --short`

Expected: no output.

Run: `git log -6 --oneline`

Expected: the six implementation commits from Tasks 1-6 appear above the design commit `ce5bdbe`.
