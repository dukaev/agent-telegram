# Codex Skill Project-Aware Onboarding Design

## Summary

Install the bundled `agent-telegram` skill into an existing project skill
directory automatically, while requiring explicit consent for any global
installation. Run onboarding only during an interactive no-argument invocation
so scripts, CI, flags, and subcommands keep their current behavior.

Codex currently discovers repository skills under `.agents/skills` from the
current working directory up to the repository root, and user skills under
`$HOME/.agents/skills`. The onboarding flow follows these canonical locations
while recognizing an existing legacy `${CODEX_HOME:-$HOME/.codex}/skills`
installation to avoid duplicates.

## Considered Approaches

### Install globally without asking

This is convenient but rejected because invoking an unrelated CLI command
would silently change the user's cross-project Codex configuration.

### Ask before every installation

This is maximally explicit but unnecessarily interrupts users when the current
project already contains a team-owned `.agents/skills` directory intended for
repo-scoped skills.

### Project-aware hybrid

This is the selected approach:

- automatically install into the nearest existing project `.agents/skills`;
- ask before installing or creating the global `$HOME/.agents/skills` target;
- never overwrite an existing `agent-telegram` entry;
- never run onboarding in non-interactive or automated invocations.

## Goals

- Make `agent-telegram` immediately available to Codex in projects that have
  already opted into repo-scoped skills.
- Keep global skill installation an explicit user decision.
- Match current Codex repository and user skill locations.
- Avoid duplicate installations when the skill already exists at any applicable
  project, canonical user, or legacy user location.
- Preserve user-owned files and require an explicit command for updates.
- Keep onboarding invisible to scripts, agents, CI, flags, and subcommands.

## Non-Goals

- Do not install from npm lifecycle hooks.
- Do not create a project `.agents/skills` directory automatically.
- Do not silently install into `$HOME/.agents/skills` even when it exists.
- Do not repair, migrate, compare, or remove existing skill contents.
- Do not automatically update an existing skill.
- Do not search arbitrary `skills/` directories; only canonical
  `.agents/skills` project locations count.

## Eligibility

Run onboarding only when all conditions are true:

1. The root command is invoked with no positional arguments and no flags.
2. Both stdin and stdout are terminals.
3. The `CI` environment variable is empty.
4. The normal root welcome text has been printed successfully.

`--help`, `--version`, global flags, subcommands, pipes, redirected output, and
CI never inspect or modify skill locations and never prompt.

## Project Discovery

Resolve the current working directory to an absolute path. Find the nearest Git
repository boundary by walking upward until a `.git` file or directory is
found. Do not invoke an external `git` process.

When a repository boundary exists, inspect `.agents/skills` in every directory
from the current working directory through the repository root, inclusive.
When no repository boundary exists, inspect only `$CWD/.agents/skills`.

Follow these rules:

1. Consider only existing directories or symlinks that resolve to directories.
2. If `agent-telegram` already exists in any discovered project skill
   directory, stop without changing anything.
3. Before installing, check the canonical and legacy user targets described
   below; if any already contains `agent-telegram`, stop without creating a
   duplicate project copy.
4. Otherwise, if one or more project skill directories exist, select the one
   nearest to the current working directory.
5. Install into `<nearest>/.agents/skills/agent-telegram` with `force=false`.
6. Report the successful automatic project installation with one concise line
   on stderr.

If a `.agents/skills` path exists but is not a directory, print a warning and
stop onboarding. Do not fall back to a global installation because the project
contains an ambiguous, potentially user-owned entry at the canonical path.

If automatic project installation fails, print a warning and stop onboarding.
Do not prompt for a global fallback.

## Global Flow

Use the canonical global target:

```text
$HOME/.agents/skills/agent-telegram
```

Reach this flow only when no project `.agents/skills` directory exists.

Before prompting, check all of the following locations for an existing
`agent-telegram` entry:

- `$HOME/.agents/skills/agent-telegram`;
- `${CODEX_HOME}/skills/agent-telegram` when `CODEX_HOME` is set;
- `$HOME/.codex/skills/agent-telegram` when `CODEX_HOME` is unset.

Treat any file, directory, or symlink at one of those targets as an existing,
user-owned installation and stop silently. The legacy check prevents a user
who installed an earlier `agent-telegram` release from receiving a duplicate.

When no existing installation is found and the user has not dismissed this
exact canonical target, write this prompt to stderr:

```text
Install the Agent Telegram skill globally for Codex?
Target: /absolute/path/to/.agents/skills/agent-telegram
This makes the skill available across projects. [y/N]
```

Only `y` and `yes`, compared case-insensitively after trimming, count as
consent. Consent may create `$HOME/.agents/skills` and installs with
`force=false`.

An empty line, `n`, `no`, or another completed response declines. EOF or a read
error makes no change and records no preference.

## Dismissal State

Store a completed global decline at:

```text
$HOME/.agent-telegram/skill-prompt-dismissed
```

The marker contains the absolute, cleaned canonical global target followed by
a newline. Use file mode `0600` and parent mode `0700`. Suppress the prompt only
when the stored target matches the current canonical global target.

Project auto-installation ignores this marker because the marker represents a
decision about global installation only.

The user can restore the global prompt by deleting the marker, or install at
any time with the explicit command:

```bash
agent-telegram skills install agent-telegram
```

## Explicit Install Command

Change the default target of `agent-telegram skills install agent-telegram` to
the canonical `$HOME/.agents/skills` directory. Keep `--target` as an explicit
override for custom or legacy locations. Do not migrate or remove existing
`${CODEX_HOME:-$HOME/.codex}/skills` content.

Keep `--force` available only on the explicit command. Onboarding never passes
`force=true`.

## Existing Skill Safety

Use `os.Lstat` when determining whether a skill target already exists, so a
dangling symlink is also treated as user-owned. Never inspect the contents of
an existing target and never replace it automatically.

When multiple project `.agents/skills` directories exist, first search all of
them for an existing `agent-telegram` entry before choosing the nearest empty
destination. This avoids creating a duplicate nested copy when a parent-level
project skill is already available.

## Components

### Skill location resolver

Add focused helpers under `internal/skills` to:

- resolve canonical and legacy user targets;
- discover the Git boundary without spawning a process;
- enumerate applicable project `.agents/skills` directories;
- detect files, directories, and symlinks safely;
- return one of three decisions: already installed, auto-install project, or
  prompt global.

### Global preference store

Keep dismissal comparison and secure marker writes separate from location
selection. The preference store only handles the canonical global target.

### Root command integration

Keep TTY, CI, flag, prompt, and response handling in the root command. The
filesystem resolver remains independent from Cobra and terminal I/O so it can
be tested with temporary directory trees.

## Data Flow

1. Cobra dispatches the eligible no-argument root command.
2. The root command prints its existing welcome text.
3. Eligibility rejects CI, flags, and non-terminal streams.
4. The resolver scans applicable project skill directories and all applicable
   canonical or legacy user targets for an existing skill.
5. Any existing project, canonical user, or legacy user skill ends onboarding
   silently.
6. An existing project skills directory triggers automatic installation into
   the nearest applicable directory.
7. Without a project skills directory, a matching global dismissal marker ends
   onboarding silently.
8. Otherwise, the root command asks for global consent.
9. Consent installs into `$HOME/.agents/skills`; decline records the global
    target; EOF records nothing.
10. All onboarding messages go to stderr and all onboarding failures preserve
    the root command's successful exit status.

## Error Handling

- Current-directory or home-directory resolution failure: skip onboarding
  silently.
- Permission or I/O failure while inspecting a discovered canonical path: print
  one warning and stop onboarding.
- Malformed project `.agents/skills` entry: warn and stop; do not fall back.
- Project installation failure: warn and stop; do not prompt globally.
- Global installation failure: warn, print the explicit install command, and
  do not create a dismissal marker.
- EOF: return without installing or recording a preference.
- Non-EOF input failure: warn and return without changing state.
- Dismissal write failure: warn and return success.
- Existing target or dangling symlink: return silently without mutation.

## Testing

Use temporary HOME, CODEX_HOME, CWD, and repository trees. Cover:

- project discovery at CWD, a parent directory, and repository root;
- nearest project directory selection;
- stopping at the nearest `.git` boundary;
- a `.git` directory and a worktree-style `.git` file;
- non-repository behavior limited to CWD;
- automatic project installation without reading stdin;
- an existing project skill in any applicable directory suppressing install;
- project files, directories, valid symlinks, and dangling symlinks remaining
  unchanged;
- malformed `.agents/skills` warning without global fallback;
- project installation failure without global fallback;
- canonical global target resolution at `$HOME/.agents/skills`;
- legacy CODEX_HOME and `$HOME/.codex` installations suppressing the prompt;
- existing canonical or legacy user installations suppressing project
  auto-installation;
- global consent for both an existing and absent `$HOME/.agents/skills`;
- global decline and target-specific dismissal;
- `y`, `yes`, uppercase consent, empty, negative, and unrecognized responses;
- EOF and read failure;
- `force=false` for project and global onboarding;
- non-terminal stdin, non-terminal stdout, CI, flags, and subcommands;
- byte-for-byte unchanged non-interactive root stdout;
- onboarding errors never changing the root exit status;
- the explicit install command defaulting to `$HOME/.agents/skills` while
  `--target` continues to work.

## Documentation

Update README, command help, manifests, generated LLM guidance, and development
notes to use the canonical project `.agents/skills` and user
`$HOME/.agents/skills` locations. Mention that existing project skill
directories enable automatic repo-scoped installation, while global
installation always requires consent.

## Acceptance Criteria

- An eligible run automatically installs into the nearest existing project
  `.agents/skills` when no applicable project copy exists.
- A project directory or target is never created outside an already existing
  `.agents/skills` directory.
- Global installation and creation occur only after `y` or `yes`.
- Existing canonical, project, and legacy targets are never modified.
- A global decline suppresses only future prompts for that canonical target and
  never blocks project auto-installation.
- CI, pipes, redirected output, flags, and subcommands never run onboarding.
- The explicit install command defaults to `$HOME/.agents/skills` and retains
  `--target` plus explicit `--force` behavior.
- Onboarding failures never prevent the root welcome command from succeeding.

## Reference

- [OpenAI Codex: Build skills — Where to save skills](https://learn.chatgpt.com/docs/build-skills#where-to-save-skills)
