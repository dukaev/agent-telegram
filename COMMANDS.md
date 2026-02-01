# Agent-Telegram Commands

CLI tool for interacting with Telegram via MTProto API.

## Server

| Command | Description |
|---------|-------------|
| `serve` | Start IPC server with Telegram client |
| `status` | Check if server is running |
| `stop` | Stop the IPC server |
| `watch` | Watch server logs in real-time |

## Authentication

| Command | Description |
|---------|-------------|
| `login` | Interactive Telegram login |
| `logout` | Logout and clear session |
| `my-info` | Get your profile information |

## Messages (`msg`)

| Command | Description |
|---------|-------------|
| `msg read` | Mark messages as read |
| `msg typing` | Send typing indicator |
| `msg scheduled` | List scheduled messages |
| `msg delete` | Delete message(s) |
| `msg forward` | Forward message |
| `msg pin` | Pin/unpin message |
| `msg reaction` | Add emoji reaction |
| `msg inspect-buttons` | View inline buttons |
| `msg press-button` | Press inline button |
| `msg inspect-keyboard` | View reply keyboard |

## Send

```bash
send --to @user "text"           # Text message
send --to @user --photo file     # Photo
send --to @user --video file     # Video
send --to @user --voice file     # Voice message
send --to @user --video-note file # Video note (circle)
send --to @user --sticker file   # Sticker
send --to @user --gif file       # GIF/animation
send --to @user --document file  # Document
send --to @user --audio file     # Audio
send --to @user --contact phone  # Contact
send --to @user --poll "Q?"      # Poll
send --to @user --latitude X --longitude Y  # Location
send --to @user --reply-to ID "text"        # Reply
```

## Chat Management (`chat`)

| Command | Description |
|---------|-------------|
| `chat list` | List all chats |
| `chat info` | Get chat information |
| `chat open` | View chat messages |
| `chat create-group` | Create group |
| `chat create-channel` | Create channel |
| `chat join` | Join via invite link |
| `chat subscribe` | Subscribe to channel |
| `chat leave` | Leave chat/channel |
| `chat invite` | Invite users |
| `chat edit-title` | Edit title |
| `chat set-photo` | Set photo |
| `chat delete-photo` | Delete photo |
| `chat pin` | Pin/unpin in list |
| `chat mute` | Mute/unmute |
| `chat archive` | Archive/unarchive |
| `chat topics` | List forum topics |
| `chat invite-link` | Get/create invite link |

### Members & Admins

| Command | Description |
|---------|-------------|
| `chat participants` | List members |
| `chat admins` | List admins |
| `chat banned` | List banned users |
| `chat promote-admin` | Promote to admin |
| `chat demote-admin` | Demote admin |

### Settings

| Command | Description |
|---------|-------------|
| `chat slow-mode` | Set slow mode (seconds) |
| `chat permissions` | Set default permissions |

## Contacts (`contact`)

| Command | Description |
|---------|-------------|
| `contact list` | List contacts |
| `contact add` | Add contact |
| `contact delete` | Delete contact |

## User (`user`)

| Command | Description |
|---------|-------------|
| `user info` | Get user info |
| `user ban` | Block/unblock user |
| `user mute` | Mute/unmute user |

## Folders (`folders`)

| Command | Description |
|---------|-------------|
| `folders list` | List chat folders |
| `folders create` | Create folder |
| `folders delete` | Delete folder |

## Privacy (`privacy`)

| Command | Description |
|---------|-------------|
| `privacy get` | Get privacy setting |
| `privacy set` | Set privacy setting |

**Privacy keys:** `status_timestamp`, `phone_number`, `profile_photo`, `forwards`, `phone_call`, `voice_messages`, `about`

**Rules:** `allow_all`, `allow_contacts`, `disallow_all`, `allow_close_friends`

## Search

| Command | Description |
|---------|-------------|
| `search` | Search messages/chats |
| `search --global` | Global search |

## Other

| Command | Description |
|---------|-------------|
| `open @user` | Open chat by username |
| `updates` | Get pending updates |
| `llms-txt` | Generate LLM documentation |

---

## IPC Methods (JSON-RPC)

All commands are also available via IPC socket at `/tmp/agent-telegram.sock`:

```bash
# Start server
agent-telegram serve

# Call method
echo '{"method":"send_message","params":{"peer":"@user","message":"Hi"}}' | nc -U /tmp/agent-telegram.sock
```

### Available Methods (77 total)

**Messages:** `send_message`, `send_reply`, `update_message`, `delete_message`, `forward_message`, `get_messages`, `clear_messages`, `clear_history`, `read_messages`, `set_typing`, `get_scheduled_messages`

**Media:** `send_photo`, `send_video`, `send_file`, `send_voice`, `send_video_note`, `send_sticker`, `send_gif`, `send_location`, `send_contact`, `send_poll`

**Reactions:** `add_reaction`, `remove_reaction`, `list_reactions`

**Buttons:** `inspect_inline_buttons`, `press_inline_button`, `inspect_reply_keyboard`

**Pins:** `pin_message`, `unpin_message`, `pin_chat`

**Chats:** `get_chats`, `get_topics`, `create_group`, `create_channel`, `edit_title`, `set_photo`, `delete_photo`, `leave`, `invite`, `join_chat`, `subscribe_channel`, `get_invite_link`

**Members:** `get_participants`, `get_admins`, `get_banned`, `promote_admin`, `demote_admin`

**Settings:** `set_slow_mode`, `set_chat_permissions`

**Folders:** `get_folders`, `create_folder`, `delete_folder`

**Users:** `get_me`, `get_user_info`, `update_profile`, `update_avatar`, `block`, `unblock`

**Contacts:** `get_contacts`, `add_contact`, `delete_contact`

**Privacy:** `get_privacy`, `set_privacy`

**Search:** `search_global`, `search_in_chat`

**Stickers:** `get_sticker_packs`
