# Thread-Aware Bot Flows Design

## Goal

Make Telegram forum-topic context a first-class part of message reads, all send operations, waits, and high-level bot flows. The CLI, JSON-RPC schema, and MTProto requests must preserve the same `threadId` and `replyTo` semantics end to end, including private bots that expose chat topics.

The work targets the current `origin/main` branch (`v0.7.16`) and ships as a separate pull request. It does not include the unrelated local chat-membership design commit.

## Compatibility

Existing commands and JSON fields keep their current behavior when `threadId` and `replyTo` are zero or omitted. The opaque `replyTo` object remains in `MessageResult` for backward compatibility while normalized scalar metadata is added alongside it.

The canonical topic commands are:

```text
agent-telegram chat topics @bot
agent-telegram bot step @bot --send "/start" --thread-id 77
agent-telegram bot press @bot 123 --text "Изменить"
agent-telegram msg list @bot --thread-id 77
agent-telegram send @bot "ping" --thread-id 77
agent-telegram send @bot --photo photo.jpg --thread-id 77
agent-telegram msg wait @bot --thread-id 77 --after-id 150
```

Existing `--to` forms and index-based `bot press` remain supported.

## Normalized Message Metadata

`types.MessageResult` gains:

```go
ThreadID         int64 `json:"threadId,omitempty"`
IsTopicMessage   bool  `json:"isTopicMessage,omitempty"`
ReplyToMessageID int64 `json:"replyToMessageId,omitempty"`
```

Conversion from `tg.MessageReplyHeader` follows these rules:

1. `ReplyToMessageID` is populated from `ReplyToMsgID` whenever it is non-zero.
2. If `ReplyToTopID` is non-zero, it becomes `ThreadID` and `IsTopicMessage` is true.
3. Otherwise, when `ForumTopic` is true and `ReplyToMsgID` is non-zero, `ReplyToMsgID` becomes `ThreadID` and `IsTopicMessage` is true.
4. Non-topic replies keep `ThreadID == 0` and `IsTopicMessage == false`.

This normalization is applied by the shared message converter, so history reads, reply-thread reads, single-message reads, waits, and bot flows see identical metadata.

## Thread-Aware Reads

`types.GetMessagesParams` gains `ThreadID int64` with JSON name `threadId`. `message.Client.GetMessages` keeps `messages.getHistory` when the field is omitted or zero. When it is non-zero, it calls `messages.getReplies` with the resolved peer, `MsgID: int(ThreadID)`, the requested limit, and the existing offset translated to `OffsetID`.

The result envelope remains `GetMessagesResult`, including its existing pagination fields, so the IPC method stays `get_messages`. `agent-telegram msg list` adds `--thread-id` and passes the value through without introducing a second user-facing read command.

## Thread-Aware Sends

All message and media send parameter structs embed one shared type:

```go
type ThreadTarget struct {
    ThreadID int64 `json:"threadId,omitempty"`
    ReplyTo  int64 `json:"replyTo,omitempty"`
}
```

This includes text, photo, file/document/audio, video, video note, location, voice, contact, poll, GIF, sticker, and dice. Existing explicit reply operations continue to work and use the same MTProto construction rules where applicable.

A shared MTProto helper constructs `tg.InputReplyToMessage` as follows:

- neither value: return `nil`;
- `ThreadID` only: set `ReplyToMsgID` to the thread root;
- `ReplyTo` only: set `ReplyToMsgID` to the replied-to message;
- both values: set `ReplyToMsgID` to `ReplyTo` and `TopMsgID` to `ThreadID`.

Every `messages.sendMessage` and `messages.sendMedia` request uses this helper. This keeps Telegram's `top_msg_id` requirement consistent for replies to non-root messages inside a topic.

The unified `send` command exposes `--thread-id` and `--reply-to` once as common flags and forwards them for every send mode. Existing specialized send subcommands that share `SendFlags` receive the same flags. A wait requested after sending inherits the selected `threadId`.

## Private Bot Topics

`chat.Client.GetTopics` removes its local `InputPeerChannel` type check and calls `messages.getForumTopics` with any resolved `tg.InputPeerClass`, including `InputPeerUser`. `chat topics` accepts either a positional peer or `--to`.

Telegram errors that indicate the peer has no forum topics are classified as a stable JSON-RPC error:

```json
{
  "type": "TOPICS_NOT_ENABLED",
  "retryable": false
}
```

The classifier and error manifest gain a dedicated code/type rather than exposing `not a channel` or a generic internal error. Genuine peer-resolution and authorization errors retain their current classifications.

## Thread-Aware Waits

The shared reply waiter accepts `threadID` and includes it in `get_messages` polling requests. A message completes the wait only when all conditions hold:

```text
message.id > afterID
message.out == false
threadID == 0 || message.threadId == threadID
```

`msg wait` adds `--thread-id`. Timeout metadata and suggested recovery commands include the thread ID when non-zero, preventing a retry from widening back to the whole chat. Send flows and `bot step` pass their selected topic through the same waiter.

## Bot Step and Button Press

`bot step` adds `--thread-id`. Its optional send, follow-up wait, latest-message read, generated next actions, and structured parameters all retain that topic context.

`bot press` supports either an index or `--text`. Text selection maps to the existing IPC `buttonText` field; the command rejects ambiguous invocations that provide conflicting selectors. Generated next actions prefer button text when it is available because keyboard indices may change after edits.

Before pressing, `bot press` reads the source message and stores a compact snapshot containing its ID, `threadId`, `editDate`, text, and normalized inline-button content. The source message's `threadId` is inherited automatically for subsequent polling and generated actions.

After the press, polling can complete with either:

1. `new_message`: a new incoming message after the source message ID in the inherited thread; or
2. `message_edited`: the source message has a changed `editDate`, text, or inline keyboard compared with the pre-press snapshot.

The successful wait metadata includes the event, and the completed message is the new or edited message:

```json
{
  "completed": true,
  "event": "message_edited",
  "message": {}
}
```

The high-level bot state remains backward compatible by continuing to expose its current `message`, `action`, `wait`, and `nextActions` fields.

## Error Handling

Negative thread and reply IDs are rejected as validation errors at the parameter or CLI boundary. A missing topic, inaccessible root message, or invalid reply target is otherwise returned from Telegram and classified through the existing RPC error pipeline.

`TOPICS_NOT_ENABLED` is non-retryable because retrying the same peer cannot enable forum topics. Wait timeouts remain `TIMEOUT`; when the button press itself succeeded, the response continues to preserve the action as a partial result.

## Testing

Unit and contract tests cover:

- message header normalization for top-level topic messages, nested topic replies, and ordinary replies;
- JSON schema exposure of `threadId` and `replyTo` across all send parameter types;
- `messages.getHistory` versus `messages.getReplies` selection and offset mapping;
- all three non-empty MTProto reply-target combinations for text and representative media builders;
- CLI registration and IPC payloads for `msg list`, `send`, `msg wait`, `bot step`, and `bot press --text`;
- `chat topics` with `InputPeerUser` and `TOPICS_NOT_ENABLED` classification/manifest metadata;
- wait filtering against outbound messages, older messages, the main chat, and neighboring topics;
- automatic thread inheritance for button presses;
- button-press completion on a new message, `editDate`, text, or inline-keyboard change;
- backward-compatible behavior when thread fields are omitted.

The final validation runs formatting, focused package tests, the full Go test suite, and any existing lint or build command documented by the repository.
