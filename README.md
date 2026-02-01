# agent-telegram

Telegram IPC agent CLI - A command-line tool for interacting with Telegram via an IPC server.

## Features

- Start a background IPC server with Telegram client
- Query chats, messages, and user info
- Interactive login with 2FA support
- Paginated results for large datasets
- JSON output for scripting

## Installation

### npm (recommended)

```bash
npm install -g agent-telegram
```

### From source

```bash
go build -o agent-telegram .
```

## Quick Start

### 1. Set up credentials

Create a `.env` file or set environment variables:

```bash
export TELEGRAM_APP_ID="your_app_id"
export TELEGRAM_APP_HASH="your_app_hash"
export TELEGRAM_PHONE="+1234567890"
```

Get your API credentials at https://my.telegram.org

### 2. Start the server

```bash
./agent-telegram serve
```

The server runs in the background and listens on `/tmp/agent-telegram.sock`

### 3. Use commands

In another terminal:

```bash
./agent-telegram get-me
./agent-telegram chats
./agent-telegram open @username
```

## Commands

### Server

```bash
./agent-telegram serve                    # Start IPC server
./agent-telegram serve -s /tmp/my.sock    # Custom socket path
```

### Query Commands

```bash
# Get current user info
./agent-telegram get-me

# List chats with pagination
./agent-telegram chats -l 20 -o 0

# Open and view messages from a user/chat
./agent-telegram open @username
./agent-telegram open @username -l 20      # 20 messages
./agent-telegram open @username -l 20 -o 50  # with offset

# Get updates (pops from store)
./agent-telegram get-updates -l 10

# Check server status
./agent-telegram status
```

### JSON Output

All commands support `--json` / `-j` flag:

```bash
./agent-telegram open @username -j | jq
```

## IPC Options

| Flag | Description | Default |
|------|-------------|---------|
| `-s, --socket` | Path to Unix socket | `/tmp/agent-telegram.sock` |

## Examples

```bash
# Start server in background
./agent-telegram serve &

# Get user info
./agent-telegram get-me

# List first 20 chats
./agent-telegram chats --limit 20

# View recent messages from @durov
./agent-telegram open @durok

# View messages with pagination, JSON output
./agent-telegram open @telegram -l 50 -o 100 -j | jq '.messages[] | .text'
```

## Development

See [DEVELOPMENT.md](DEVELOPMENT.md) for:
- Architecture overview
- Adding new commands
- Common patterns
- Project structure

## License

MIT
