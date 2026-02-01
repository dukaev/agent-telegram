# Development Guide

## Architecture Overview

agent-telegram uses a **two-process architecture**:

```
┌─────────────────┐         IPC          ┌──────────────────┐        MTProto
│   CLI Process   │ ◄──────────────────► │  Server Process  │ ◄────────────────► Telegram
│                 │    Unix Socket       │                  │     (gotd/td)
│  cmd/*.go       │    JSON-RPC 2.0      │  internal/ipc/   │
│  internal/      │                      │  telegram/       │
│  cliutil/       │                      │                  │
└─────────────────┘                      └──────────────────┘
        │                                        │
        ▼                                        ▼
   Short-lived                             Long-running
   (per command)                           (background daemon)
```

### Process Communication

```
CLI Command
    ↓
cliutil.Runner (auto-starts server if needed)
    ↓
internal/ipc.Client (connects to socket)
    ↓
JSON-RPC Request → /tmp/agent-telegram.sock → internal/ipc.SocketServer
    ↓
internal/ipc.Server (method dispatch)
    ↓
Handler (internal/telegram/ipc/)
    ↓
telegram.Client.* (domain operation)
    ↓
JSON-RPC Response ← Unix Socket ← Result
```

---

## Project Structure

```
agent-telegram/
├── cmd/                           # CLI commands
│   ├── auth/                      # login, logout
│   ├── chat/                      # 26+ chat subcommands
│   ├── contact/                   # add, list, delete
│   ├── folders/                   # folder management
│   ├── get/                       # data retrieval
│   ├── message/                   # msg subcommands
│   ├── mute/                      # mute operations
│   ├── open/                      # open resources
│   ├── privacy/                   # privacy settings
│   ├── search/                    # search functionality
│   ├── send/                      # unified send command
│   ├── sys/                       # status, llms-txt
│   ├── user/                      # user operations
│   ├── root.go                    # root command, groups
│   ├── register.go                # command registration
│   ├── serve.go                   # server startup
│   └── stop.go                    # server shutdown
│
├── internal/
│   ├── ipc/                       # IPC infrastructure
│   │   ├── client.go              # JSON-RPC client
│   │   ├── server.go              # JSON-RPC server
│   │   ├── socket.go              # Unix socket server
│   │   ├── types.go               # Request/Response types
│   │   ├── interface.go           # Client interfaces
│   │   └── methods.go             # ping, echo methods
│   │
│   ├── telegram/ipc/              # Telegram RPC handlers
│   │   ├── register.go            # 95+ method handlers
│   │   ├── handler.go             # Generic handler factory
│   │   ├── handlers.go            # Handler implementations
│   │   └── *.go                   # Specific handlers
│   │
│   ├── cliutil/                   # CLI utilities
│   │   ├── runner.go              # Command runner with auto-start
│   │   ├── recipient.go           # @user/ID normalization
│   │   ├── pagination.go          # Limit/offset helper
│   │   ├── listcmd.go             # List command pattern
│   │   ├── togglecmd.go           # Toggle command pattern
│   │   ├── print.go               # Output formatting
│   │   └── filter.go              # Item filtering
│   │
│   ├── paths/                     # File path management
│   │   └── paths.go               # Config dir, logs, PID, lock
│   │
│   ├── config/                    # Configuration loading
│   ├── auth/                      # Auth service
│   ├── tgauth/                    # Telegram auth flow
│   └── types/                     # Shared types
│
├── telegram/                      # Telegram client
│   ├── client.go                  # Main client wrapper
│   ├── accessors.go               # Domain client accessors
│   ├── domain_interfaces.go       # Domain interfaces
│   ├── updates.go                 # Update handling
│   ├── peer.go                    # Peer types
│   ├── chat/                      # Chat operations
│   ├── message/                   # Message operations
│   ├── media/                     # Media handling
│   ├── user/                      # User operations
│   ├── search/                    # Search client
│   ├── pin/                       # Pin operations
│   ├── reaction/                  # Reactions
│   ├── types/                     # Domain types
│   └── helpers/                   # Utilities
│
├── cli/                           # Interactive UI
│   ├── ui/                        # Login UI
│   ├── steps/                     # Auth steps (phone, code, password)
│   └── components/                # UI components
│
├── main.go                        # Entry point
├── go.mod
├── install.sh                     # curl installer
└── package.json                   # npm package config
```

---

## Key Components

### 1. IPC Layer (`internal/ipc/`)

**Protocol**: JSON-RPC 2.0 over Unix domain sockets

```go
// Request format
{
  "jsonrpc": "2.0",
  "method": "send_message",
  "params": {"peer": "@user", "message": "Hello"},
  "id": 1
}

// Response format
{
  "jsonrpc": "2.0",
  "result": {"id": 123, "peer": "@user"},
  "id": 1
}
```

**Error codes**:

| Code | Meaning |
|------|---------|
| -32700 | Parse error |
| -32600 | Invalid request |
| -32601 | Method not found |
| -32602 | Invalid params |
| -32603 | Internal error |
| -32001 | Server not running |
| -32002 | Not authorized |
| -32003 | Not initialized |

### 2. CLI Utilities (`internal/cliutil/`)

**Runner** - Command execution with auto-start:

```go
runner := cliutil.NewRunnerFromCmd(cmd, jsonFlag)
result := runner.Call("method", params)  // Auto-starts server
runner.PrintResult(result, formatter)
```

**Recipient** - Peer identifier normalization:

```go
var to cliutil.Recipient
cmd.Flags().Var(&to, "to", "Recipient (@user or ID)")

// Usage
to.Peer()           // Returns "@username" or "123456"
to.AddToParams(m)   // Adds {"peer": "..."} to map
```

**Pagination**:

```go
pag := cliutil.NewPagination(cmd, 10, 100)  // default, max
params["limit"] = pag.Limit
params["offset"] = pag.Offset
```

### 3. Telegram Client (`telegram/`)

**Domain-driven design** with separate clients:

```go
type Client struct {
    // gotd/td client
    client *telegram.Client

    // Domain clients (lazy init)
    message  *message.Client
    chat     *chat.Client
    user     *user.Client
    media    *media.Client
    // ...
}

// Accessors
func (c *Client) Message() *message.Client
func (c *Client) Chat() *chat.Client
func (c *Client) User() *user.Client
```

### 4. Handler Registration (`internal/telegram/ipc/`)

**Generic handler factory**:

```go
// Type-safe handler
func Handler[T any, R any](
    fn func(context.Context, T) (R, error),
) HandlerFunc

// File validation handler
func FileHandler[T any, R any](
    getFile func(T) string,
    fn func(context.Context, T) (R, error),
) HandlerFunc
```

**Registration pattern**:

```go
var methodHandlers = map[string]func(Client) HandlerFunc{
    "send_message":   sendMessageHandler,
    "get_chats":      getChatsHandler,
    "delete_message": deleteMessageHandler,
    // ... 95+ handlers
}

func RegisterHandlers(srv ipc.MethodRegistrar, client Client) {
    for method, factory := range methodHandlers {
        registerHandler(srv, method, factory(client))
    }
}
```

---

## Adding a New Command

### Step 1: Add Telegram Client Method

**File**: `telegram/<domain>/client.go` or new file

```go
// telegram/message/client.go

type SendStickerParams struct {
    Peer string
    File string
}

type SendStickerResult struct {
    ID   int64  `json:"id"`
    Peer string `json:"peer"`
}

func (c *Client) SendSticker(ctx context.Context, p SendStickerParams) (*SendStickerResult, error) {
    if c.api == nil {
        return nil, ErrNotInitialized
    }

    inputPeer, err := c.parent.ResolvePeer(ctx, p.Peer)
    if err != nil {
        return nil, fmt.Errorf("resolve peer: %w", err)
    }

    // Telegram API call
    result, err := c.api.MessagesSendMedia(ctx, &tg.MessagesSendMediaRequest{
        Peer:  inputPeer,
        Media: &tg.InputMediaUploadedDocument{...},
    })
    if err != nil {
        return nil, err
    }

    return &SendStickerResult{
        ID:   extractMessageID(result),
        Peer: p.Peer,
    }, nil
}
```

### Step 2: Add IPC Handler

**File**: `internal/telegram/ipc/handlers.go` (or new file)

```go
type sendStickerParams struct {
    Peer string `json:"peer"`
    File string `json:"file"`
}

func sendStickerHandler(client Client) HandlerFunc {
    return FileHandler(
        func(p sendStickerParams) string { return p.File },
        func(ctx context.Context, p sendStickerParams) (*message.SendStickerResult, error) {
            return client.Message().SendSticker(ctx, message.SendStickerParams{
                Peer: p.Peer,
                File: p.File,
            })
        },
    )
}
```

### Step 3: Register Handler

**File**: `internal/telegram/ipc/register.go`

```go
var methodHandlers = map[string]func(Client) HandlerFunc{
    // ... existing handlers
    "send_sticker": sendStickerHandler,  // Add here
}
```

### Step 4: Add CLI Command

**File**: `cmd/send/send.go` (extend) or new file

```go
// cmd/send/sticker.go
package send

import (
    "github.com/spf13/cobra"
    "agent-telegram/internal/cliutil"
)

var stickerFile string

var sendStickerCmd = &cobra.Command{
    Use:   "sticker",
    Short: "Send a sticker",
    Run: func(cmd *cobra.Command, args []string) {
        runner := sendFlags.NewRunner()

        params := map[string]any{
            "file": stickerFile,
        }
        sendFlags.To.AddToParams(params)

        result := runner.Call("send_sticker", params)
        runner.PrintResult(result, func(r any) {
            cliutil.FormatSuccess(r, "send_sticker")
        })
    },
}

func init() {
    SendCmd.AddCommand(sendStickerCmd)
    sendFlags.Register(sendStickerCmd)
    sendStickerCmd.Flags().StringVar(&stickerFile, "file", "", "Sticker file path")
    sendStickerCmd.MarkFlagRequired("file")
}
```

### Step 5: Build and Test

```bash
# Build
go build -o agent-telegram .

# Lint
golangci-lint run

# Test
go test ./...

# Manual test
./agent-telegram send sticker --to @user --file sticker.webp
```

---

## Common Patterns

### Pagination

```go
var cmdLimit, cmdOffset int

func init() {
    cmd.Flags().IntVarP(&cmdLimit, "limit", "l", 20, "Number of items")
    cmd.Flags().IntVarP(&cmdOffset, "offset", "o", 0, "Offset")
}

// In Run function
pag := cliutil.NewPagination(cmd, 20, 100)
params := map[string]any{
    "limit":  pag.Limit,
    "offset": pag.Offset,
}
```

### JSON Output

```go
var jsonOutput bool

cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "JSON output")

// In Run function
runner := cliutil.NewRunnerFromCmd(cmd, jsonOutput)
result := runner.Call("method", params)
runner.PrintResult(result, humanFormatter)
```

### Toggle Commands (pin/unpin, mute/unmute)

```go
// Use cliutil.ToggleCmd pattern
toggle := cliutil.NewToggleCmd("pin", "unpin", "Pin or unpin chat")
toggle.Register(parentCmd, func(enabled bool, args []string) {
    method := "pin_chat"
    if !enabled {
        method = "unpin_chat"
    }
    // ...
})
```

### Recipient Flag

```go
var to cliutil.Recipient

cmd.Flags().Var(&to, "to", "Recipient (@username or chat ID)")

// In Run function
if to.IsEmpty() {
    fmt.Fprintln(os.Stderr, "Error: --to is required")
    os.Exit(1)
}
params := make(map[string]any)
to.AddToParams(params)  // Adds "peer" key
```

---

## Server Lifecycle

### Startup (`cmd/serve.go`)

1. Load and validate credentials
2. Fork to background (unless `-f`)
3. Acquire lock file (`~/.agent-telegram/server.lock`)
4. Write PID file (`~/.agent-telegram/server.pid`)
5. Setup structured logging (`~/.agent-telegram/server.log`)
6. Start Telegram client (with retry logic)
7. Start IPC server on Unix socket
8. Handle signals for graceful shutdown

### Auto-Start (`internal/cliutil/runner.go`)

Commands automatically start the server:

1. Check if server running (status RPC)
2. If not running, acquire start lock
3. Fork server process
4. Wait up to 30 seconds for startup
5. Release lock
6. Proceed with command

### Shutdown (`cmd/stop.go`)

1. Try graceful shutdown via RPC
2. If fails, read PID from file
3. Send SIGTERM
4. Wait for process exit
5. Clean up PID file

---

## File Locations

| File | Path | Purpose |
|------|------|---------|
| Socket | `/tmp/agent-telegram.sock` | IPC communication |
| Session | `~/.agent-telegram/session.json` | Telegram auth state |
| Log | `~/.agent-telegram/server.log` | Server logs (JSON) |
| PID | `~/.agent-telegram/server.pid` | Running server PID |
| Lock | `~/.agent-telegram/server.lock` | Instance lock (flock) |

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `TELEGRAM_APP_ID` | Telegram API App ID (optional, has default) |
| `TELEGRAM_APP_HASH` | Telegram API App Hash (optional, has default) |
| `TELEGRAM_PHONE` | Phone number for auth |
| `AGENT_TELEGRAM_APP_ID` | Alternative for APP_ID |
| `AGENT_TELEGRAM_APP_HASH` | Alternative for APP_HASH |
| `AGENT_TELEGRAM_SESSION_PATH` | Custom session path |

---

## Testing

### Manual Testing

```bash
# Start server in foreground (see logs)
./agent-telegram serve -f

# In another terminal
./agent-telegram status
./agent-telegram chat list -l 5
./agent-telegram send --to @user "Test"
```

### IPC Testing

```bash
# Direct JSON-RPC call
echo '{"jsonrpc":"2.0","method":"ping","id":1}' | nc -U /tmp/agent-telegram.sock

# With jq
echo '{"jsonrpc":"2.0","method":"get_me","id":1}' | nc -U /tmp/agent-telegram.sock | jq
```

### Unit Tests

```bash
go test ./...
go test ./internal/ipc/...
go test ./telegram/...
```

---

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/gotd/td` | Telegram MTProto client |
| `github.com/spf13/cobra` | CLI framework |
| `github.com/charmbracelet/bubbletea` | Interactive TUI |
| `github.com/charmbracelet/bubbles` | TUI components |
| `github.com/knadh/koanf/v2` | Configuration |
| `github.com/joho/godotenv` | .env loading |

---

## Common Gotchas

### 1. Access Hash Required

When resolving usernames, always use the full peer resolution:

```go
inputPeer, err := client.ResolvePeer(ctx, "@username")
// NOT just the username string
```

### 2. Message ID Types

Telegram uses `int`, JSON uses `float64`:

```go
// When extracting from map[string]any
id := int64(m["id"].(float64))
```

### 3. Response Variants

Telegram API returns variant types:

```go
switch m := messagesClass.(type) {
case *tg.MessagesMessages:
    return m.Messages, m.Users, nil
case *tg.MessagesMessagesSlice:
    return m.Messages, m.Users, nil
case *tg.MessagesChannelMessages:
    return m.Messages, m.Users, nil
}
```

### 4. Peer Types

Always handle all peer types:

- `*tg.PeerUser` - Direct messages
- `*tg.PeerChat` - Legacy groups
- `*tg.PeerChannel` - Channels and supergroups

### 5. Handler Panics

Handlers have panic recovery - nil pointer → `ErrNotInitialized`:

```go
// This is safe - will return error instead of crash
result, err := client.Message().Send(ctx, params)
```

---

## Release Process

1. Update version in `package.json`
2. Create git tag: `git tag v0.1.0`
3. Push tag: `git push origin v0.1.0`
4. GoReleaser builds binaries for all platforms
5. npm publish (if applicable)
