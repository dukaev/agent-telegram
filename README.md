# agent-telegram

Telegram automation CLI for AI agents. Fast Go binary with MTProto API.

By Aslan Dukaev [X](https://x.com/dukaev) ·[Telegram](https://t.me/dukaev)

## Installation

### Package Managers

```bash
bun add -g agent-telegram
npm install -g agent-telegram
```

### From Source

```bash
git clone https://github.com/dukaev/agent-telegram
cd agent-telegram
go build -o agent-telegram .
```

## Quick Start

```bash
agent-telegram login                      # Interactive login
agent-telegram my-info                    # Get your profile
agent-telegram chat list                  # List all chats
agent-telegram chat open @username        # View messages from chat
agent-telegram send --to @user "Hello!"   # Send message
agent-telegram stop                       # Stop server
```

## Commands

### Authentication

```bash
agent-telegram login                      # Interactive login
agent-telegram logout                     # Logout and clear session
agent-telegram my-info                    # Get your profile information
agent-telegram llms-txt                   # Generate full CLI documentation for LLMs

```

### Send Messages

```bash
agent-telegram send --to @user "text"               # Text message
agent-telegram send --to @user --photo image.png    # Photo
agent-telegram send --to @user --video video.mp4    # Video
agent-telegram send --to @user --voice voice.ogg    # Voice message
agent-telegram send --to @user --video-note vid.mp4 # Video note (circle)
agent-telegram send --to @user --sticker file.webp  # Sticker
agent-telegram send --to @user --gif anim.mp4       # GIF/animation
agent-telegram send --to @user --document file.pdf  # Document
agent-telegram send --to @user --audio music.mp3    # Audio
agent-telegram send --to @user --contact "+1234567890" --first-name "John"  # Contact
agent-telegram send --to @user --reply-to 123 "Reply text"                  # Reply
agent-telegram send --to @user --poll "Question?" --option "Yes" --option "No"  # Poll
agent-telegram send --to @user --latitude 55.7558 --longitude 37.6173       # Location
```

### Message Management (`msg`)

```bash
agent-telegram msg read @user             # Mark messages as read
agent-telegram msg typing @user           # Send typing indicator
agent-telegram msg scheduled @user        # List scheduled messages
agent-telegram msg delete @user 123       # Delete message by ID
agent-telegram msg forward @user --to @other 123  # Forward message
agent-telegram msg pin @user 123          # Pin message
agent-telegram msg pin @user 123 --unpin  # Unpin message
agent-telegram msg reaction @user 123 "👍" # Add reaction
agent-telegram msg inspect-buttons @user 123      # View inline buttons
agent-telegram msg press-button @user 123 0 0     # Press button (row, col)
agent-telegram msg inspect-keyboard @user         # View reply keyboard
```

### Chat Management (`chat`)

```bash
agent-telegram chat list                  # List all chats
agent-telegram chat list -l 50            # List with limit
agent-telegram chat info @channel         # Get chat information
agent-telegram chat open @user            # View messages
agent-telegram chat open @user -l 50      # View 50 messages
agent-telegram chat open @user -l 50 -o 100   # With offset
agent-telegram chat create-group "Name" @user1 @user2   # Create group
agent-telegram chat create-channel "Name" "Description" # Create channel
agent-telegram chat join https://t.me/+invite           # Join via link
agent-telegram chat subscribe @channel    # Subscribe to channel
agent-telegram chat leave @group          # Leave chat/channel
agent-telegram chat invite @group @user   # Invite user
agent-telegram chat edit-title @group "New Title"       # Edit title
agent-telegram chat set-photo @group photo.jpg          # Set photo
agent-telegram chat delete-photo @group   # Delete photo
agent-telegram chat pin @group            # Pin chat in list
agent-telegram chat pin @group --unpin    # Unpin from list
agent-telegram chat mute @group           # Mute notifications
agent-telegram chat mute @group --unmute  # Unmute
agent-telegram chat archive @group        # Archive chat
agent-telegram chat archive @group --unarchive  # Unarchive
agent-telegram chat topics @forum         # List forum topics
agent-telegram chat invite-link @group    # Get/create invite link
```

### Members & Admins

```bash
agent-telegram chat participants @group   # List members
agent-telegram chat admins @group         # List admins
agent-telegram chat banned @group         # List banned users
agent-telegram chat promote-admin @group @user    # Promote to admin
agent-telegram chat demote-admin @group @user     # Demote admin
agent-telegram chat slow-mode @group 30   # Set slow mode (seconds)
agent-telegram chat permissions @group    # Set default permissions
```

### Contacts (`contact`)

```bash
agent-telegram contact list               # List contacts
agent-telegram contact add "+1234567890" "John" "Doe"   # Add contact
agent-telegram contact delete @user       # Delete contact
```

### User (`user`)

```bash
agent-telegram user info @user            # Get user info
agent-telegram user ban @user             # Block user
agent-telegram user ban @user --unban     # Unblock user
agent-telegram user mute @user            # Mute user
agent-telegram user mute @user --unmute   # Unmute user
```

### Folders (`folders`)

```bash
agent-telegram folders list               # List chat folders
agent-telegram folders create "Work"      # Create folder
agent-telegram folders delete 1           # Delete folder by ID
```

### Privacy (`privacy`)

```bash
agent-telegram privacy get phone_number   # Get privacy setting
agent-telegram privacy set phone_number allow_contacts  # Set privacy
```

**Privacy keys:** `status_timestamp`, `phone_number`, `profile_photo`, `forwards`, `phone_call`, `voice_messages`, `about`

**Rules:** `allow_all`, `allow_contacts`, `disallow_all`, `allow_close_friends`

### Search

```bash
agent-telegram search "query"             # Search in chats
agent-telegram search "query" --global    # Global search
agent-telegram search "query" --in @user  # Search in specific chat
```

### Server

```bash
agent-telegram serve                      # Start IPC server (background)
agent-telegram status                     # Check server status
agent-telegram stop                       # Stop server
```

### Other

```bash
agent-telegram open @user                 # Quick open chat (alias)
agent-telegram updates                    # Get pending updates
agent-telegram updates -l 50              # Get 50 updates
```

## Options

| Option | Description |
|--------|-------------|
| `-s, --socket <path>` | Path to Unix socket (default: `/tmp/agent-telegram.sock`) |
| `-j, --json` | JSON output (for agents) |
| `-l, --limit <n>` | Limit results |
| `-o, --offset <n>` | Offset for pagination |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `TELEGRAM_APP_ID` | Telegram API App ID (optional, has default) |
| `TELEGRAM_APP_HASH` | Telegram API App Hash (optional, has default) |
| `TELEGRAM_PHONE` | Phone number for auth (optional) |
| `AGENT_TELEGRAM_SESSION_PATH` | Custom session file path |

Default API credentials are built-in, so you can start using agent-telegram immediately. To use your own credentials, get them at https://my.telegram.org and set via environment variables or `.env` file.

## Agent Mode

Use `--json` / `-j` for machine-readable output:

```bash
agent-telegram chat list --json
agent-telegram chat open @user -l 10 --json
agent-telegram send --to @user "Hello" --json
```

### Optimal AI Workflow

```bash
# 1. Start server and verify status
agent-telegram serve
agent-telegram status --json

# 2. List chats to find targets
agent-telegram chat list -l 20 --json

# 3. Read messages from a chat
agent-telegram chat open @username -l 50 --json

# 4. Send messages
agent-telegram send --to @username "Hello!" --json

# 5. Check for new messages
agent-telegram updates --json
```

## IPC Protocol (JSON-RPC)

All commands communicate via JSON-RPC over Unix socket at `/tmp/agent-telegram.sock`:

```bash
# Direct IPC call
echo '{"method":"send_message","params":{"peer":"@user","message":"Hi"}}' | nc -U /tmp/agent-telegram.sock
```

### Available Methods (77+)

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

**System:** `status`, `shutdown`, `ping`

## Architecture

agent-telegram uses a client-daemon architecture:

```
┌─────────────┐      IPC       ┌──────────────┐      MTProto
│ CLI Command │ ─────────────> │ IPC Server   │ ──────────────> Telegram
│ (Go binary) │  Unix Socket   │ (background) │   (gotd/td)
└─────────────┘                └──────────────┘
```

- **CLI Commands** - Parse arguments, communicate with daemon via IPC
- **IPC Server** - Background daemon managing Telegram connection
- **Telegram Client** - MTProto client using [gotd/td](https://github.com/gotd/td) library

The daemon starts automatically on first command and persists between commands for fast subsequent operations.

### File Locations

| File | Path |
|------|------|
| Unix socket | `/tmp/agent-telegram.sock` |
| Session | `~/.agent-telegram/session.json` |
| Logs | `~/.agent-telegram/server.log` |
| PID file | `~/.agent-telegram/server.pid` |
| Lock file | `~/.agent-telegram/server.lock` |

## Sessions

Each session maintains its own:
- Telegram connection
- Authentication state
- Message history cache
- Update store

Use `--socket` to run multiple isolated instances:

```bash
# Different sessions
agent-telegram --socket /tmp/agent1.sock serve
agent-telegram --socket /tmp/agent2.sock serve

# Use specific session
agent-telegram --socket /tmp/agent1.sock chat list
```

## Usage with AI Agents

### Just ask the agent

The simplest approach - just tell your agent to use it:

```
Use agent-telegram to send a message to @username. Run agent-telegram --help to see available commands.
```

```markdown
## Telegram Automation

Use `agent-telegram` for Telegram automation. Run `agent-telegram --help` for all commands.

Core workflow:
1. `agent-telegram serve` - Start background server
2. `agent-telegram status` - Verify connection
3. `agent-telegram chat list --json` - List available chats
4. `agent-telegram chat open @user --json` - Read messages
5. `agent-telegram send --to @user "message"` - Send message
```

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md) for:
- Architecture overview
- Adding new commands
- Common patterns
- Project structure

## License

MIT
