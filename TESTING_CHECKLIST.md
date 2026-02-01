# Testing Checklist

**Test Chat:** `@user`
**Date:** 2026-02-01
**Tester:** Claude

---

## 1. Authentication

- [ ] 1.1 `login` — Interactive login
- [ ] 1.2 `logout` — Logout and clear session
- [ ] 1.3 `my-info` — Get current user profile

---

## 2. Server

- [x] 2.1 `serve` — Start server (background)
- [ ] 2.2 `serve -f` — Start server (foreground)
- [x] 2.3 `status` — Check server status
- [x] 2.4 `stop` — Stop server gracefully
- [ ] 2.5 `stop --force` — Force stop server

---

## 3. Send Messages

- [x] 3.1 `send --to @user "Hello"` — Text message
- [x] 3.2 `send --to @user --photo test.png` — Photo
- [x] 3.3 `send --to @user --video test.mov` — Video
- [x] 3.4 `send --to @user --audio test.m4a` — Audio
- [x] 3.5 `send --to @user --document test.txt` — Document
- [x] 3.6 `send --to @user --voice test.m4a` — Voice message
- [x] 3.7 `send --to @user --video-note test.mov` — Video note
- [x] 3.8 `send --to @user --sticker test.png` — Sticker
- [x] 3.9 `send --to @user --gif test.gif` — GIF
- [x] 3.10 `send --to @user --file test.png` — Auto-detect type
- [x] 3.11 `send --to @user --photo test.png --caption "Cap"` — With caption
- [x] 3.12 `send --to @user --reply-to <id> "Reply"` — Reply
- [x] 3.13 `send --to @bot --poll "Q?" --option "A" --option "B"` — Poll (not in Saved Messages)
- [x] 3.14 `send --to @user --latitude 55.75 --longitude 37.62` — Location
- [x] 3.15 `send --to @user --contact "+123" --first-name "John"` — Contact
- [x] 3.16 `send update --to @user --message-id <id> "Edited"` — Edit

---

## 4. Message Operations

- [x] 4.1 `msg read --to @user` — Mark as read
- [x] 4.2 `msg read --to @user --max-id <id>` — Mark up to ID
- [x] 4.3 `msg typing --to @user` — Typing indicator
- [x] 4.4 `msg scheduled --to @user` — Scheduled messages
- [x] 4.5 `msg delete --to @user --message-id <id>` — Delete
- [x] 4.6 `msg forward --to @user --from <peer> --message-id <id>` — Forward
- [x] 4.7 `msg pin --to @user --message-id <id>` — Pin
- [x] 4.8 `msg pin --to @user --message-id <id> --disable` — Unpin
- [x] 4.9 `msg inspect-buttons --to @bot --message-id <id>` — Buttons
- [x] 4.10 `msg press-button --to @bot --message-id <id> <row> <col>` — Press
- [x] 4.11 `msg inspect-keyboard --to @user` — Keyboard
- [x] 4.12 `msg reaction <id> "emoji" --to @user` — Add reaction
- [x] 4.13 `msg reaction <id> "emoji" --to @user --big` — Big reaction

---

## 5. Chat — Info & List

- [x] 5.1 `chat list` — List all chats
- [x] 5.2 `chat list --limit 5` — With limit
- [x] 5.3 `chat list --search "query"` — Filter by name
- [x] 5.4 `chat list --type channel` — Filter by type
- [x] 5.5 `chat info --to @user` — Chat info
- [x] 5.6 `chat open @user` — Get messages
- [x] 5.7 `chat open @user --limit 5` — With limit
- [x] 5.8 `chat topics --to -100XXXXXXXXX` — Forum topics

---

## 6. Chat — Actions

- [x] 6.1 `chat archive --to @user` — Archive
- [x] 6.2 `chat archive --to @user --disable` — Unarchive
- [x] 6.3 `chat mute --to @user` — Mute
- [x] 6.4 `chat mute --to @user --disable` — Unmute
- [x] 6.5 `chat pin --to @user` — Pin chat
- [x] 6.6 `chat pin --to @user --disable` — Unpin chat

---

## 7. Chat — Join/Leave

- [x] 7.1 `chat join --inviteLink <link>` — Join via invite
- [x] 7.2 `chat subscribe --channel @channel` — Subscribe
- [x] 7.3 `chat leave --to @chat` — Leave

---

## 8. Chat — Management

- [x] 8.1 `chat create-group --title "Test" --members @user1` — Create group
- [ ] 8.2 `chat create-channel --title "Test"` — Create channel (FLOOD_WAIT)
- [x] 8.3 `chat edit-title --to @supergroup --title "New"` — Edit title
- [x] 8.4 `chat set-photo --to <chat> --file photo.jpg` — Set photo
- [x] 8.5 `chat delete-photo --to @supergroup` — Delete photo
- [x] 8.6 `chat slow-mode --to @supergroup --seconds 30` — Set slow mode
- [x] 8.7 `chat slow-mode --to @supergroup --seconds 0` — Disable slow mode
- [x] 8.8 `chat permissions --to @supergroup` — Permissions
- [x] 8.9 `chat invite --to @supergroup --members @username` — Invite user
- [x] 8.10 `chat invite-link --to @supergroup` — Get invite link

---

## 9. Chat — Admin

- [x] 9.1 `chat participants --to @channel` — List participants
- [x] 9.2 `chat participants --to @channel --limit 10` — With limit
- [x] 9.3 `chat admins --to @channel` — List admins
- [x] 9.4 `chat banned --to @channel` — List banned
- [x] 9.5 `chat promote-admin --to @supergroup --user @user` — Promote
- [x] 9.6 `chat demote-admin --to @supergroup --user @user` — Demote

---

## 10. Root Commands

- [x] 10.1 `chats` — List chats
- [x] 10.2 `chats --limit 5` — With limit
- [x] 10.3 `chats --search "query"` — Filter
- [x] 10.4 `chats --type user` — By type
- [x] 10.5 `open @user` — Open messages
- [x] 10.6 `open @user --limit 5` — With limit
- [x] 10.7 `open <invite_link>` — Join via invite
- [x] 10.8 `updates` — Get updates
- [x] 10.9 `updates --limit 10` — With limit

---

## 11. Search

- [x] 11.1 `search global "query"` — Global search
- [x] 11.2 `search global "query" --limit 10` — With limit
- [x] 11.3 `search in-chat --to @user "query"` — In chat
- [x] 11.4 `search in-chat --to @user "query" --limit 5` — With limit

---

## 12. Contacts

- [x] 12.1 `contact list` — List contacts
- [x] 12.2 `contact add --phone "+123" --first-name "Test"` — Add
- [x] 12.3 `contact add --phone "+123" --first-name "T" --last-name "U"` — With last name
- [x] 12.4 `contact delete --username @username` — Delete

---

## 13. User Operations

- [x] 13.1 `user info @username` — User info
- [x] 13.2 `user info` — Own info
- [x] 13.3 `user ban --to @username` — Block
- [x] 13.4 `user ban --to @username --disable` — Unblock
- [x] 13.5 `user mute --to @user` — Mute
- [x] 13.6 `user mute --to @user --disable` — Unmute

---

## 14. Folders

- [x] 14.1 `folders list` — List folders
- [x] 14.2 `folders create --title "Test" --include-contacts` — Create
- [x] 14.3 `folders delete --id <id>` — Delete

---

## 15. Privacy

- [x] 15.1 `privacy get --key <key>` — Get setting
- [x] 15.2 `privacy set --key <key> --rule <rule>` — Set

---

## 16. Utility

- [x] 16.1 `llms-txt` — Generate LLM docs
- [x] 16.2 `--help` — Show help
- [x] 16.4 `--version` — Show version
- [x] 16.5 `<cmd> --dry-run` — Dry run
- [x] 16.6 `chat open @user --json` — JSON output
- [x] 16.7 `send ... -q` — Quiet mode

---

## Summary

| # | Category | Total | Pass | Fail | Skip |
|---|----------|-------|------|------|------|
| 1 | Authentication | 3 | 0 | 0 | 3 |
| 2 | Server | 5 | 3 | 0 | 2 |
| 3 | Send Messages | 16 | 16 | 0 | 0 |
| 4 | Message Ops | 13 | 13 | 0 | 0 |
| 5 | Chat Info/List | 8 | 8 | 0 | 0 |
| 6 | Chat Actions | 6 | 6 | 0 | 0 |
| 7 | Chat Join/Leave | 3 | 3 | 0 | 0 |
| 8 | Chat Management | 10 | 9 | 0 | 1 |
| 9 | Chat Admin | 6 | 6 | 0 | 0 |
| 10 | Root Commands | 9 | 9 | 0 | 0 |
| 11 | Search | 4 | 4 | 0 | 0 |
| 12 | Contacts | 4 | 4 | 0 | 0 |
| 13 | User Ops | 6 | 6 | 0 | 0 |
| 14 | Folders | 3 | 3 | 0 | 0 |
| 15 | Privacy | 2 | 2 | 0 | 0 |
| 16 | Utility | 6 | 6 | 0 | 0 |
| | **TOTAL** | **104** | **98** | **0** | **6** |

---

## Issues Fixed During Testing

| # | Test | Issue | Fix |
|---|------|-------|-----|
| 1 | 9.1 | `filter is nil` in participants | Added `Filter: &tg.ChannelParticipantsRecent{}` |
| 2 | 11.3 | `filter is nil` in search | Added `Filter: &tg.InputMessagesFilterEmpty{}` |
| 3 | 12.1 | Flag conflict `-s` | Changed to `-S` for search |
| 4 | 8.6 | Flag conflict `-s` | Changed to `-S` for seconds |
| 5 | 6.1-6.4 | Methods not implemented | Implemented Archive/Mute |
| 6 | All | Numeric IDs not working | Fixed ResolvePeer + CLI Recipient |
| 7 | 3.8 | `stickerset is nil` | Added `InputStickerSetEmpty{}` |
| 8 | 8.4 | `set_photo` not implemented | Implemented with file upload |

---

## Notes

- Commands with supergroup require a public group or channel (not basic group)
- Poll cannot be sent to Saved Messages (Telegram limitation)
- `chat set-photo` requires minimum 512x512 image
- Some admin commands require admin rights in the chat
- Media tests (photo, video, audio, etc.) require actual files
