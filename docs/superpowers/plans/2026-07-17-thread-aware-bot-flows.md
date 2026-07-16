# Thread-Aware Bot Flows Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Carry Telegram topic context through normalized messages, CLI and IPC reads/sends/waits, private-bot topic discovery, and edit-aware bot button flows.

**Architecture:** Extend existing JSON types without changing RPC method names, route non-zero `threadId` reads through `messages.getReplies`, and use one MTProto reply-target builder for all message/media sends. Add an edit-aware bot waiter that combines thread-filtered message polling with snapshots of the pressed message.

**Tech Stack:** Go 1.24, Cobra, JSON-RPC IPC, gotd/td MTProto, `tgmock`, standard `testing`.

## Global Constraints

- Base the PR on `origin/main` (`v0.7.16`) and exclude unrelated local commits.
- Preserve behavior when `threadId` and `replyTo` are omitted or zero.
- Preserve opaque `MessageResult.replyTo` while adding normalized topic fields.
- Apply `ThreadTarget` to text, photo, file/document/audio, video, video note, location, voice, contact, poll, GIF, sticker, and dice; exclude gift operations.
- Keep index-based `bot press` compatible while adding `--text`.
- Classify disabled topics as non-retryable `TOPICS_NOT_ENABLED`.

---

### Task 1: Normalize topic metadata and add thread-aware reads

**Files:**
- Modify: `telegram/types/types.go`
- Modify: `telegram/message/convert.go`
- Modify: `telegram/message/messages.go`
- Modify: `cmd/message/list.go`
- Test: `telegram/message/message_test.go`
- Test: `cmd/message/message_test.go`

**Interfaces:**
- Consumes: `tg.MessageReplyHeader` and `GetMessagesParams.ThreadID`.
- Produces: normalized `MessageResult` fields and `msg list --thread-id`.

- [ ] **Step 1: Write failing normalization tests**

Add table cases to `telegram/message/message_test.go`:

```go
tests := []struct {
    name string
    header *tg.MessageReplyHeader
    threadID int64
    topic bool
    replyTo int64
}{
    {"nested topic", &tg.MessageReplyHeader{ReplyToMsgID: 88, ReplyToTopID: 77, ForumTopic: true}, 77, true, 88},
    {"topic root", &tg.MessageReplyHeader{ReplyToMsgID: 77, ForumTopic: true}, 77, true, 77},
    {"ordinary reply", &tg.MessageReplyHeader{ReplyToMsgID: 42}, 0, false, 42},
}
```

Pass each header through the existing converter and assert `ThreadID`, `IsTopicMessage`, and `ReplyToMessageID`.

- [ ] **Step 2: Run the test and confirm failure**

Run: `go test ./telegram/message -run 'Test.*Topic.*Metadata' -count=1`

Expected: compile failure because the fields do not exist.

- [ ] **Step 3: Add the result fields and conversion rules**

Add to `MessageResult`:

```go
ThreadID         int64 `json:"threadId,omitempty"`
IsTopicMessage   bool  `json:"isTopicMessage,omitempty"`
ReplyToMessageID int64 `json:"replyToMessageId,omitempty"`
```

After preserving `r.ReplyTo = convertReplyHeader(header)`, add:

```go
if header.ReplyToMsgID != 0 { r.ReplyToMessageID = int64(header.ReplyToMsgID) }
if header.ReplyToTopID != 0 {
    r.ThreadID, r.IsTopicMessage = int64(header.ReplyToTopID), true
} else if header.ForumTopic && header.ReplyToMsgID != 0 {
    r.ThreadID, r.IsTopicMessage = int64(header.ReplyToMsgID), true
}
```

- [ ] **Step 4: Write failing history/replies dispatch tests**

Capture both request types with `tgmock`. Assert zero thread calls `MessagesGetHistoryRequest`; assert `{ThreadID: 77, Limit: 20, Offset: 100}` calls:

```go
&tg.MessagesGetRepliesRequest{Peer: peer, MsgID: 77, OffsetID: 100, Limit: 20}
```

Also register the message CLI and require `ListCmd.Flags().Lookup("thread-id")`.

- [ ] **Step 5: Implement read routing and CLI payload**

Add this field to `GetMessagesParams`:

```go
ThreadID int64 `json:"threadId,omitempty"`
```

In `GetMessages`, reject negative values, then select:

```go
var messagesClass tg.MessagesMessagesClass
if params.ThreadID != 0 {
    messagesClass, err = c.API().MessagesGetReplies(ctx, &tg.MessagesGetRepliesRequest{
        Peer: inputPeer, MsgID: int(params.ThreadID), OffsetID: params.Offset, Limit: params.Limit,
    })
} else {
    messagesClass, err = c.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
        Peer: inputPeer, OffsetID: params.Offset, Limit: params.Limit,
    })
}
```

Register `--thread-id` in `cmd/message/list.go` and add `threadId` to the IPC map only when non-zero.

- [ ] **Step 6: Verify and commit**

Run: `go test ./telegram/message ./telegram/types ./cmd/message -count=1`

Expected: PASS.

```bash
git add telegram/types/types.go telegram/message/convert.go telegram/message/messages.go telegram/message/message_test.go cmd/message/list.go cmd/message/message_test.go
git commit -m "feat: add thread-aware message reads"
```

### Task 2: Apply a common thread target to all message sends

**Files:**
- Create: `telegram/internal/replytarget/replytarget.go`
- Create: `telegram/internal/replytarget/replytarget_test.go`
- Modify: `telegram/types/types_send.go`
- Modify: `telegram/types/types_message.go`
- Modify: `telegram/message/send.go`
- Modify: `telegram/media/photo.go`
- Modify: `telegram/media/document.go`
- Modify: `telegram/media/location.go`
- Modify: `telegram/media/voice.go`
- Modify: `telegram/media/contact.go`
- Modify: `telegram/media/poll.go`
- Modify: `telegram/media/sticker.go`
- Modify: `telegram/media/dice.go`
- Test: `telegram/message/message_test.go`
- Test: `telegram/media/media_test.go`
- Test: `telegram/types/params_test.go`

**Interfaces:**
- Produces: `types.ThreadTarget` and `replytarget.Build(types.ThreadTarget)`.
- Consumes: the helper from every `MessagesSendMessageRequest` and `MessagesSendMediaRequest`.

- [ ] **Step 1: Write failing helper and JSON tests**

Test these exact cases:

```go
tests := []struct {
    in types.ThreadTarget
    wantReply, wantTop int
    wantNil bool
}{
    {wantNil: true},
    {in: types.ThreadTarget{ThreadID: 77}, wantReply: 77},
    {in: types.ThreadTarget{ReplyTo: 88}, wantReply: 88},
    {in: types.ThreadTarget{ThreadID: 77, ReplyTo: 88}, wantReply: 88, wantTop: 77},
}
```

JSON-marshal each messaging `Send*Params` type with `ThreadID: 77, ReplyTo: 88` and assert both camel-case keys exist. Assert `SendStarGiftParams` is unchanged.

- [ ] **Step 2: Confirm failure**

Run: `go test ./telegram/internal/replytarget ./telegram/types -count=1`

Expected: compile failure because the package and type do not exist.

- [ ] **Step 3: Add the shared type and builder**

Embed this type in every messaging send parameter and remove the duplicate dice field:

```go
type ThreadTarget struct {
    ThreadID int64 `json:"threadId,omitempty"`
    ReplyTo  int64 `json:"replyTo,omitempty"`
}
```

Embed `ThreadTarget` in `SendReplyParams` as well; its existing `MessageID` remains the required replied-to message ID while `ThreadID` supplies the topic root.

Create:

```go
package replytarget

import (
    "agent-telegram/telegram/types"
    "github.com/gotd/td/tg"
)

func Build(target types.ThreadTarget) *tg.InputReplyToMessage {
    if target.ThreadID == 0 && target.ReplyTo == 0 { return nil }
    replyID := target.ReplyTo
    if replyID == 0 { replyID = target.ThreadID }
    reply := &tg.InputReplyToMessage{ReplyToMsgID: int(replyID)}
    if target.ThreadID != 0 && target.ReplyTo != 0 { reply.SetTopMsgID(int(target.ThreadID)) }
    return reply
}
```

- [ ] **Step 4: Use the builder in every MTProto request**

Import `telegram/internal/replytarget` in message/media builders and set:

```go
ReplyTo: replytarget.Build(params.ThreadTarget),
```

Change the document helper to preserve target data:

```go
func (c *Client) SendDocument(ctx context.Context, peer, file, mimeType, caption string, target types.ThreadTarget) (*types.SendFileResult, error)
```

Pass `params.ThreadTarget` from file/video. Keep captions/entities intact. For legacy `SendReply`, construct `types.ThreadTarget{ThreadID: params.ThreadID, ReplyTo: params.MessageID}` so nested replies set `TopMsgID`.

- [ ] **Step 5: Add request-capture tests**

For text and every media sender, capture the request and assert:

```go
reply, ok := req.ReplyTo.(*tg.InputReplyToMessage)
if !ok || reply.ReplyToMsgID != 88 || reply.TopMsgID != 77 {
    t.Fatalf("reply target = %#v", req.ReplyTo)
}
```

- [ ] **Step 6: Verify and commit**

Run: `go test ./telegram/internal/replytarget ./telegram/types ./telegram/message ./telegram/media -count=1`

Expected: PASS.

```bash
git add telegram/internal/replytarget telegram/types/types_send.go telegram/types/types_message.go telegram/types/params_test.go telegram/message telegram/media
git commit -m "feat: send messages to forum topics"
```

### Task 3: Carry thread context through send and wait CLIs

**Files:**
- Modify: `cmd/send/flags.go`
- Modify: `cmd/send/send.go`
- Modify: `cmd/send/text.go`
- Modify: `cmd/send/media.go`
- Modify: `cmd/send/location.go`
- Modify: `cmd/send/contact.go`
- Modify: `cmd/send/poll.go`
- Modify: `cmd/send/dice.go`
- Modify: `cmd/send/wait.go`
- Modify: `cmd/message/wait.go`
- Test: `cmd/send/send_test.go`
- Test: `cmd/send/wait_test.go`
- Test: `cmd/message/message_test.go`

**Interfaces:**
- Produces: common `--thread-id`/`--reply-to`, thread-filtered waits, and thread-preserving recovery commands.
- Consumes: normalized `message.threadId` and `get_messages.threadId`.

- [ ] **Step 1: Write failing CLI and wait-filter tests**

Require both flags on unified send and every send subcommand. Set `ThreadID: 77, ReplyTo: 88` in each `buildSendParams` branch and assert both keys survive.

Build a wait result containing outbound, old, main-chat, thread-78, and thread-77 messages. Assert `findReply(result, 100, 77)` returns only the incoming thread-77 message; assert thread zero retains whole-chat behavior.

- [ ] **Step 2: Confirm failure**

Run: `go test ./cmd/send ./cmd/message -run 'Test.*(Send|Wait|Reply|Surface)' -count=1`

Expected: missing flags/signatures.

- [ ] **Step 3: Add common send flags**

Extend `SendFlags` with `ThreadID` and `ReplyTo`, register both flags in all registration paths, and add:

```go
func (f *SendFlags) AddThreadTarget(params map[string]any) {
    if f.ThreadID != 0 { params["threadId"] = f.ThreadID }
    if f.ReplyTo != 0 { params["replyTo"] = f.ReplyTo }
}
```

Call it from every unified/specialized payload. Replace command-local reply globals. For `send_reply`, map `ReplyTo` to existing `messageId` and keep `threadId`. Reject negative IDs before IPC.

- [ ] **Step 4: Change waiter signatures and filtering**

Use:

```go
func WaitForReply(poller ReplyPoller, peer string, threadID, afterMsgID int64, timeout time.Duration) WaitOutcome
func findReply(result any, afterMsgID, threadID int64) map[string]any
```

Add `ThreadID` to `WaitOutcome`, include `threadId` in polling only when non-zero, and filter with:

```go
if msgID <= afterMsgID { continue }
if threadID != 0 && cliutil.ExtractInt64(msg, "threadId") != threadID { continue }
return msg
```

Carry thread context through `HandleWaitReply`, success/timeout metadata, audit details, and suggested recovery commands.

- [ ] **Step 5: Add `msg wait --thread-id`**

Register and validate the flag, pass it to `WaitForReply`, and include it in returned wait metadata. Change all send callers to pass their selected thread.

- [ ] **Step 6: Verify and commit**

Run: `go test ./cmd/send ./cmd/message -count=1`

Expected: PASS.

```bash
git add cmd/send cmd/message/wait.go cmd/message/message_test.go
git commit -m "feat: make send waits thread aware"
```

### Task 4: Support private bot topics and typed errors

**Files:**
- Modify: `telegram/chat/topics.go`
- Modify: `cmd/chat/topics.go`
- Modify: `internal/ipc/errors.go`
- Modify: `internal/ipc/httpserver_support.go`
- Modify: `internal/telegram/ipc/register.go`
- Test: `telegram/chat/chat_test.go`
- Test: `cmd/chat/chat_test.go`
- Test: `internal/telegram/ipc/register_test.go`
- Test: `cmd/sys/manifest_test.go`

**Interfaces:**
- Produces: `chat topics [peer]` for any `InputPeer`, plus `TOPICS_NOT_ENABLED`.

- [ ] **Step 1: Write failing private-peer and classifier tests**

Use a resolved `&tg.InputPeerUser{UserID: 1, AccessHash: 2}` and assert `MessagesGetForumTopicsRequest.Peer` preserves it. Add classifier cases for `CHAT_FORUM_MISSING` and `FORUM_NOT_ENABLED`. Require the command to accept one positional peer.

- [ ] **Step 2: Confirm failure**

Run: `go test ./telegram/chat ./cmd/chat ./internal/telegram/ipc ./cmd/sys -run 'Test.*(Topics|Error|Manifest)' -count=1`

Expected: `not a channel` and missing typed error/positional syntax.

- [ ] **Step 3: Remove local peer-type rejection and support both CLI forms**

Delete the `InputPeerChannel` switch. Change the command to `Use: "topics [peer]"`, `Args: cobra.MaximumNArgs(1)`, make `--to` optional, populate it from `args[0]`, and validate that either form supplied a peer.

- [ ] **Step 4: Add stable error metadata and classification**

Add:

```go
ErrCodeTopicsNotEnabled = -32013
ErrorTypeTopicsNotEnabled = "TOPICS_NOT_ENABLED"
```

Add it to `ErrorTypesManifest` with `retryable: false`. Classify structured Telegram names and lower-case message forms before generic forbidden/internal cases.

- [ ] **Step 5: Verify and commit**

Run: `go test ./telegram/chat ./cmd/chat ./internal/ipc ./internal/telegram/ipc ./cmd/sys -count=1`

Expected: PASS.

```bash
git add telegram/chat/topics.go telegram/chat/chat_test.go cmd/chat/topics.go cmd/chat/chat_test.go internal/ipc/errors.go internal/ipc/httpserver_support.go internal/telegram/ipc/register.go internal/telegram/ipc/register_test.go cmd/sys/manifest_test.go
git commit -m "feat: support private bot topics"
```

### Task 5: Make bot step/press thread- and edit-aware

**Files:**
- Modify: `cmd/bot/bot.go`
- Create: `cmd/bot/wait.go`
- Modify: `cmd/bot/bot_test.go`
- Modify: `telegram/message/inline.go`
- Modify: `telegram/message/message_test.go`

**Interfaces:**
- Produces: `bot step --thread-id`, `bot press --text`, inherited thread context, and events `new_message`/`message_edited`.
- Consumes: `get_message`, `get_messages`, `inspect_inline_buttons`, and `press_inline_button`.

- [ ] **Step 1: Write failing surface, selector, and snapshot tests**

Require `StepCmd --thread-id` and `PressCmd --text`. Test index selection, text selection, missing selector, and conflicting selectors. Test that changes to `editDate`, text, or buttons each mark a source snapshot changed while an identical message does not.

- [ ] **Step 2: Confirm failure**

Run: `go test ./cmd/bot ./telegram/message -run 'Test.*(Bot|Press|Snapshot|Inline)' -count=1`

Expected: missing flags/helpers.

- [ ] **Step 3: Add snapshots and edit-aware polling**

Create:

```go
type messageSnapshot struct {
    ID, ThreadID, EditDate int64
    Text, Buttons string
}

type botWaitOutcome struct {
    Message map[string]any
    Event string
    Polls int
    Completed bool
}
```

Canonicalize buttons with `json.Marshal`. `snapshotChanged` compares matching IDs and detects any change to edit date, text, or canonical buttons. `waitForBotEvent` polls thread-aware `get_messages` for a new incoming message and `get_message` for the original ID each cycle, returning `new_message` or `message_edited`.

- [ ] **Step 4: Thread `bot step` end to end**

Register `stepThreadID`; include it in send, wait, latest-message read, bot state, and generated next-action commands/params.

- [ ] **Step 5: Add button text and automatic thread inheritance**

Change use to `press [peer] <message_id> [button_index]` and register `pressText`. Before pressing:

```go
source := getMessage(runner, pressTo.Peer(), messageID)
snapshot := newMessageSnapshot(source)
threadID := snapshot.ThreadID
```

Send either `buttonText` or `buttonIndex` to IPC. Wait using the inherited thread. On success include:

```go
map[string]any{
    "completed": true,
    "event": outcome.Event,
    "message": outcome.Message,
    "threadId": threadID,
}
```

Generated press next actions prefer `--text` with shell-safe button text and fall back to index only when text is unavailable.

- [ ] **Step 6: Verify and commit**

Run: `go test ./cmd/bot ./cmd/send ./telegram/message -count=1`

Expected: PASS.

```bash
git add cmd/bot telegram/message/inline.go telegram/message/message_test.go
git commit -m "feat: make bot button flows thread and edit aware"
```

### Task 6: Validate contracts and publish the draft PR

**Files:**
- Modify: `cmd/registry_contract_test.go`
- Modify: `internal/docs/docs.go`
- Verify: every changed Go/Markdown file

**Interfaces:**
- Produces: complete CLI/schema/error documentation and a draft PR targeting `main`.

- [ ] **Step 1: Add final contract assertions**

Assert generated surfaces contain `chat topics [peer]`, `bot step --thread-id`, `bot press --text`, `msg list --thread-id`, `send --thread-id`, `send --reply-to`, `msg wait --thread-id`, and `TOPICS_NOT_ENABLED`. Assert `threadId` appears in message/send schemas but not gift schemas.

- [ ] **Step 2: Format and check the diff**

Run `gofmt -w` on the exact changed Go files, then run `git diff --check`.

Expected: no output from `git diff --check`.

- [ ] **Step 3: Run focused validation**

```bash
go test ./telegram/types ./telegram/internal/replytarget ./telegram/message ./telegram/media ./telegram/chat ./internal/ipc ./internal/telegram/ipc ./cmd/message ./cmd/send ./cmd/bot ./cmd/chat ./cmd/sys -count=1
```

Expected: PASS.

- [ ] **Step 4: Run repository-wide validation**

```bash
go test ./... -count=1
go vet ./...
```

Expected: PASS.

- [ ] **Step 5: Commit contract and agent-doc adjustments**

```bash
git add cmd/registry_contract_test.go internal/docs/docs.go
git commit -m "test: cover thread-aware bot flow contracts"
```

- [ ] **Step 6: Inspect and publish**

```bash
git status -sb
git diff --stat origin/main...HEAD
git log --oneline origin/main..HEAD
gh auth status
git push -u origin codex/thread-aware-bot-flows
```

Open a draft PR to `main` titled `feat: add thread-aware bot flows`. The body must summarize normalized metadata, thread reads/sends/waits, private bot topics, edit-aware button presses, and exact validation results.
