---
name: agent-telegram
description: "Use when Codex needs to operate Telegram through the agent-telegram CLI: authenticate with Telegram, start or inspect the local IPC server, list/read/send messages, interact with Telegram bots, press inline/reply keyboard buttons, inspect chats/users/gifts, or debug agent-telegram runs with audit/log/trace commands."
---

# Agent Telegram

## Core Workflow

Use the local `agent-telegram` binary for Telegram automation. Prefer the
agent-friendly contract for all nontrivial work:

```bash
RUN_ID="run_$(date +%s)"
agent-telegram server ensure --agent --run-id "$RUN_ID"
agent-telegram status --agent --run-id "$RUN_ID"
```

For command discovery, prefer generated docs from the installed binary:

```bash
agent-telegram llms-txt
agent-telegram manifest
agent-telegram <command> --schema
```

## Authentication

Authentication is QR-only and runs in a local browser:

```bash
agent-telegram auth --agent --run-id "$RUN_ID"
```

## Reading And Sending

Use compact output controls to conserve context:

```bash
agent-telegram chat list --agent --run-id "$RUN_ID" --max-items 10
agent-telegram msg list @user --agent --run-id "$RUN_ID" --verbosity compact --max-items 5 --max-text-chars 160
agent-telegram send --to @user "message" --agent --run-id "$RUN_ID"
```

Use `--dry-run --agent` before destructive, paid, or ambiguous actions.
Check `safety` in `manifest` or `--schema`; confirm with the user before
`destructive` or `paid` operations.

## Bot Flows

Prefer high-level bot commands instead of manually stitching button/message
steps:

```bash
agent-telegram bot step @bot --send "/start" --agent --run-id "$RUN_ID"
agent-telegram bot press @bot <message_id> <button_index> --agent --run-id "$RUN_ID"
agent-telegram msg wait @bot --after-id <message_id> --timeout 20s --agent --run-id "$RUN_ID"
```

Use structured `nextActions` from bot responses to decide whether to send text,
press a button, wait for a reply, or stop.

## Observability

Keep one `RUN_ID` across related commands. When something fails, inspect the
run or trace before retrying blindly:

```bash
agent-telegram run inspect "$RUN_ID" --agent
agent-telegram trace inspect <traceId> --agent --run-id "$RUN_ID"
agent-telegram audit --run-id "$RUN_ID"
agent-telegram logs --kind server --run-id "$RUN_ID"
```

Audit/log output is redacted by default. Do not request full secrets or raw
message dumps unless the user explicitly asks and the task requires it.
