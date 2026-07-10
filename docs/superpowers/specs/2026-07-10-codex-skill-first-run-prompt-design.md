# Codex Skill First-Run Prompt Design

## Summary

Offer to install the bundled `agent-telegram` Codex skill during the first
interactive invocation of `agent-telegram` with no arguments. Installation
must require explicit consent, must never overwrite an existing skill, and
must not affect scripts, pipes, CI, or normal subcommands.

## Goals

- Make the bundled Codex skill discoverable at the moment a person first
  explores the CLI.
- Install the skill with one explicit confirmation instead of requiring the
  user to discover and type a separate command.
- Preserve the current behavior for non-interactive callers and every
  subcommand.
- Avoid repeated prompts after the user declines.
- Preserve user-owned files and require an explicit command for updates.

## Non-Goals

- Do not install the skill from npm lifecycle hooks.
- Do not silently install or update a skill.
- Do not validate, repair, migrate, or compare the contents of an existing
  skill directory.
- Do not add automatic version tracking for installed skills.
- Do not change `agent-telegram skills install agent-telegram` or its `--force`
  behavior.

## User Experience

The offer is eligible only when all of the following are true:

1. The root command is invoked with no positional arguments and no flags.
2. Both stdin and stdout are terminals.
3. The `CI` environment variable is empty.
4. `${CODEX_HOME:-~/.codex}/skills/agent-telegram` does not exist.
5. The user has not dismissed the offer for the current target directory.

The existing welcome text is printed first. The prompt is then written to
stderr so stdout retains its existing command output:

```text
Install the Agent Telegram skill for Codex?
Target: /absolute/path/to/skills/agent-telegram
This adds CLI usage instructions for Codex. [y/N]
```

Input is trimmed and compared case-insensitively. Only `y` and `yes` count as
consent. An empty line, `n`, `no`, or any other completed response counts as a
decline. EOF or a read error ends the prompt without installing or recording a
decision.

After consent, the CLI calls the existing installation path with
`force=false`. Success is reported on stderr and the command exits normally.
An installation failure is reported as a warning on stderr, does not create a
dismissal marker, and does not make the root command fail.

After a decline, the CLI records the decision and exits normally. If the
decision cannot be recorded, it prints a warning on stderr but still succeeds.

## Dismissal State

Store the dismissal marker at:

```text
~/.agent-telegram/skill-prompt-dismissed
```

The file contains the absolute, cleaned default skill target directory followed
by a newline and uses mode `0600`; its parent directory uses mode `0700`.
Suppress the prompt only when the stored path equals the current target. This
allows the prompt to appear again if `CODEX_HOME` later points somewhere else.

The user can restore the prompt by deleting the marker. Regardless of the
marker, the user can always install explicitly:

```bash
agent-telegram skills install agent-telegram
```

## Existing Skill Safety

Treat any existing filesystem entry at the target path as user-owned and do not
prompt or mutate it. This includes directories, files, and symlinks. The first-
run flow never passes `force=true`. Updating or replacing an installed skill
therefore remains an explicit user action:

```bash
agent-telegram skills install agent-telegram --force
```

## Components

### Root command integration

The root command keeps responsibility for deciding whether the invocation is
eligible and for performing terminal I/O. The prompt runs after the welcome
text so the original output remains intact.

### Skill onboarding helpers

Add focused helpers under `internal/skills` for:

- resolving the default skill target;
- detecting an existing target;
- resolving and comparing dismissal state;
- recording a dismissal securely;
- installing through the existing `Install` function.

Keep prompt rendering and input handling separate from filesystem decisions so
both can be tested without a real terminal.

## Data Flow

1. Cobra dispatches the no-argument root command.
2. The root command prints the existing welcome text.
3. Eligibility checks reject flagged, non-terminal, or CI invocations.
4. The onboarding helper resolves the target and dismissal marker.
5. An existing target or matching dismissal marker ends the flow silently.
6. The root command displays the offer and reads one response.
7. Consent calls `skills.Install("agent-telegram", "", false)`.
8. Decline records the resolved target in the marker.
9. All onboarding messages go to stderr; onboarding failures leave the root
   command's exit status successful.

## Error Handling

- Failure to resolve the home directory or target: skip the offer silently,
  because onboarding must not make the CLI unusable.
- Failure to inspect the target for reasons other than non-existence: skip the
  offer and print a concise warning.
- EOF or input read failure: make no change and return success.
- Installation failure: print the error and the explicit install command, then
  return success.
- Dismissal write failure: print a warning, make no installation change, and
  return success.
- Existing target: make no attempt to determine ownership or correctness and
  return silently.

## Testing

Add table-driven unit tests covering:

- `y` and `yes` install the bundled skill;
- uppercase consent is accepted;
- empty, `n`, `no`, and unrecognized input decline;
- a decline writes a mode-`0600` marker containing the normalized target;
- a matching marker suppresses future offers;
- a marker for a different `CODEX_HOME` target does not suppress the offer;
- an existing target suppresses the offer without modifying it;
- installation never uses force and cannot overwrite an existing skill;
- non-terminal stdin or stdout suppresses the offer;
- a non-empty `CI` value suppresses the offer;
- root flags and all subcommands suppress the offer;
- EOF, target inspection failure, install failure, and marker write failure do
  not change the root command exit status;
- non-interactive root output remains byte-for-byte compatible with the current
  welcome text.

Use temporary HOME and CODEX_HOME directories in all filesystem tests. Do not
read or write the developer's real Codex skill directory.

## Documentation

Update the README quick start and generated agent guidance to mention that the
interactive root command may offer skill installation. Keep the explicit
`agent-telegram skills install agent-telegram` command documented for automated
and manual setups.

## Acceptance Criteria

- A first interactive no-argument invocation offers installation once.
- No filesystem change occurs without an affirmative `y` or `yes`, except the
  documented dismissal marker after a completed negative response.
- Existing skill paths are never changed by onboarding.
- Scripts, pipes, CI, flagged root calls, and subcommands never prompt.
- Declining suppresses only the current resolved target.
- Onboarding failures do not prevent the CLI welcome from completing
  successfully.
- The existing explicit install and force-update commands continue to work.
