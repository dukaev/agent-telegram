# Development Guide

## Architecture Overview

```
┌─────────────┐      IPC       ┌──────────────┐      Telegram API
│ CLI Command │ ──────────────> │ IPC Server   │ ──────────────────> Telegram
│  (cmd/*.go) │  Unix Socket   │ (internal/)  │   (gotd/td)
└─────────────┘                └──────────────┘
                                      │
                                      ▼
                              ┌──────────────┐
                              │   Telegram   │
                              │   Client     │
                              │ (telegram/)  │
                              └──────────────┘
```

### Components

| Layer | Location | Responsibility |
|-------|----------|-----------------|
| CLI | `cmd/*.go` | User-facing commands, argument parsing, output formatting |
| IPC Client | `internal/ipc/client.go` | JSON-RPC client for communicating with server |
| IPC Server | `internal/ipc/server.go` | JSON-RPC server running on Unix socket |
| Handlers | `internal/telegram/ipc/*.go` | Request handlers that bridge IPC to Telegram client |
| Client | `telegram/*.go` | Telegram API wrapper using gotd/td library |

---

## Adding a New Command

### Pattern: CLI Command → IPC Handler → Client Method

To add a new command (e.g., `send-message`), follow these steps:

---

### Step 1: Add Client Method

**File:** `telegram/client.go` or create `telegram/client_<feature>.go`

```go
// SendMessageParams holds parameters for sending a message.
type SendMessageParams struct {
    Peer    string
    Message string
}

// SendMessageResult is the result of SendMessage.
type SendMessageResult struct {
    ID        int64  `json:"id"`
    Date      int64  `json:"date"`
    Message   string `json:"message"`
}

// SendMessage sends a message to a peer.
func (c *Client) SendMessage(ctx context.Context, params SendMessageParams) (*SendMessageResult, error) {
    if c.client == nil {
        return nil, fmt.Errorf("client not initialized")
    }

    api := c.client.API()

    // 1. Resolve peer (username → InputPeer)
    inputPeer, err := c.resolveUsername(ctx, api, params.Peer)
    if err != nil {
        return nil, err
    }

    // 2. Call Telegram API
    result, err := api.MessagesSendMessage(ctx, &tg.MessagesSendMessageRequest{
        Peer:    inputPeer,
        Message: params.Message,
        // ... other fields
    })
    if err != nil {
        return nil, err
    }

    // 3. Return structured result
    return &SendMessageResult{
        ID:      result.ID,
        Message: params.Message,
    }, nil
}
```

---

### Step 2: Update Client Interface

**File:** `internal/telegram/ipc/getme.go`

Add the method signature to the `Client` interface:

```go
// Client is an interface for Telegram operations.
type Client interface {
    GetMe(ctx context.Context) (*tg.User, error)
    GetChats(ctx context.Context, limit, offset int) ([]map[string]interface{}, error)
    GetUpdates(limit int) []telegram.StoredUpdate
    GetMessages(ctx context.Context, params telegram.GetMessagesParams) (*telegram.GetMessagesResult, error)
    SendMessage(ctx context.Context, params telegram.SendMessageParams) (*telegram.SendMessageResult, error)  // NEW
}
```

---

### Step 3: Create IPC Handler

**File:** Create `internal/telegram/ipc/sendmessage.go`

```go
// Package ipc provides Telegram IPC handlers.
package ipc

import (
    "context"
    "encoding/json"
    "fmt"

    "agent-telegram/telegram"
)

// SendMessageParams represents parameters for send_message request.
type SendMessageParams struct {
    Peer    string `json:"peer"`
    Message string `json:"message"`
}

// SendMessageHandler returns a handler for send_message requests.
func SendMessageHandler(client Client) func(json.RawMessage) (interface{}, error) {
    return func(params json.RawMessage) (interface{}, error) {
        var p SendMessageParams
        if err := json.Unmarshal(params, &p); err != nil {
            return nil, fmt.Errorf("invalid params: %w", err)
        }

        // Validate
        if p.Peer == "" {
            return nil, fmt.Errorf("peer is required")
        }
        if p.Message == "" {
            return nil, fmt.Errorf("message is required")
        }

        result, err := client.SendMessage(context.Background(), telegram.SendMessageParams{
            Peer:    p.Peer,
            Message: p.Message,
        })
        if err != nil {
            return nil, fmt.Errorf("failed to send message: %w", err)
        }

        return result, nil
    }
}
```

---

### Step 4: Register Handler

**File:** `internal/telegram/ipc/register.go`

```go
func RegisterHandlers(srv ipc.MethodRegistrar, client Client) {
    // ... existing handlers ...

    srv.Register("send_message", func(params json.RawMessage) (interface{}, *ipc.ErrorObject) {
        result, err := SendMessageHandler(client)(params)
        if err != nil {
            return nil, &ipc.ErrorObject{
                Code:    -32000,
                Message: err.Error(),
            }
        }
        return result, nil
    })
}
```

---

### Step 5: Create CLI Command

**File:** Create `cmd/send_message.go`

```go
// Package cmd provides CLI commands.
package cmd

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/spf13/cobra"

    "agent-telegram/internal/ipc"
)

var (
    sendMessageJSON bool
)

// sendMessageCmd represents the send-message command.
var sendMessageCmd = &cobra.Command{
    Use:   "send-message @peer <message>",
    Short: "Send a message to a Telegram peer",
    Long:  `Send a message to a Telegram user or chat by username.`,
    Args:  cobra.MinimumNArgs(2),
    Run:   runSendMessage,
}

func init() {
    rootCmd.AddCommand(sendMessageCmd)
    sendMessageCmd.Flags().BoolVarP(&sendMessageJSON, "json", "j", false, "Output as JSON")
}

func runSendMessage(_ *cobra.Command, args []string) {
    socketPath, _ := rootCmd.Flags().GetString("socket")
    peer := args[0]
    message := args[1]

    client := ipc.NewClient(socketPath)
    result, err := client.Call("send_message", map[string]any{
        "peer":    peer,
        "message": message,
    })
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if sendMessageJSON {
        data, _ := json.MarshalIndent(result, "", "  ")
        fmt.Println(string(data))
    } else {
        fmt.Printf("Message sent successfully!\n")
    }
}
```

---

### Step 6: Build and Test

```bash
# Build
go build -o agent-telegram .

# Run tests (if you have them)
go test ./...

# Run linter
golangci-lint run

# Test command
./agent-telegram send-message @username "Hello, world!"
```

---

## Common Patterns

### Pagination (Limit/Offset)

```go
// CLI command flags
var (
    cmdLimit  int
    cmdOffset int
)

func init() {
    cmd.Flags().IntVarP(&cmdLimit, "limit", "l", 10, "Number of items (max 100)")
    cmd.Flags().IntVarP(&cmdOffset, "offset", "o", 0, "Offset for pagination")
}

// Validate
if cmdLimit < 1 { cmdLimit = 1 }
if cmdLimit > 100 { cmdLimit = 100 }
if cmdOffset < 0 { cmdOffset = 0 }
```

### JSON Output

```go
if cmdJSON {
    data, err := json.MarshalIndent(result, "", "  ")
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Println(string(data))
} else {
    // Human-readable output
    printHumanReadable(result)
}
```

### Resolving Username to InputPeer

```go
func (c *Client) resolveUsername(ctx context.Context, api *tg.Client, username string) (tg.InputPeerClass, error) {
    peerClass, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
        Username: strings.TrimPrefix(username, "@"),
    })
    if err != nil {
        return nil, err
    }

    switch p := peerClass.Peer.(type) {
    case *tg.PeerUser:
        return &tg.InputPeerUser{
            UserID:     p.UserID,
            AccessHash: getAccessHash(peerClass, p.UserID),
        }, nil
    case *tg.PeerChat:
        return &tg.InputPeerChat{ChatID: p.ChatID}, nil
    case *tg.PeerChannel:
        return &tg.InputPeerChannel{
            ChannelID:  p.ChannelID,
            AccessHash: getAccessHash(peerClass, p.ChannelID),
        }, nil
    default:
        return nil, fmt.Errorf("unsupported peer type: %T", p)
    }
}
```

---

## File Structure

```
agent-telegram/
├── cmd/                          # CLI commands
│   ├── root.go                  # Root command, global flags
│   ├── serve.go                 # Start IPC server
│   ├── get_me.go                # Example: simple command
│   ├── chats.go                 # Example: pagination
│   └── open.go                  # Example: with username argument
│
├── internal/
│   ├── ipc/                     # IPC infrastructure
│   │   ├── client.go           # JSON-RPC client
│   │   ├── server.go           # JSON-RPC server
│   │   └── socket.go           # Unix socket wrapper
│   │
│   └── telegram/ipc/            # Telegram IPC handlers
│       ├── register.go         # Handler registration
│       ├── getme.go            # get_me handler
│       ├── chats.go            # get_chats handler
│       ├── messages.go         # get_messages handler
│       └── [new_handler].go    # Your new handler
│
├── telegram/                     # Telegram client
│   ├── client.go               # Main client, auth
│   ├── client_dialogs.go       # Dialog/chat operations
│   ├── client_updates.go       # Update handling
│   └── [client_new].go         # Your new operations
│
├── .env                         # Credentials (TELEGRAM_APP_ID, etc.)
├── go.mod
└── DEVELOPMENT.md              # This file
```

---

## Key Dependencies

- **gotd/td** - Telegram MTProto client library
- **cobra** - CLI framework
- **godotenv** - Environment variable loading

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `TELEGRAM_APP_ID` | Telegram API App ID (from my.telegram.org) |
| `TELEGRAM_APP_HASH` | Telegram API App Hash (from my.telegram.org) |
| `TELEGRAM_PHONE` | Phone number for auth |
| `AGENT_TELEGRAM_SESSION_PATH` | Path to session file (optional) |

---

## Common Gotchas

1. **Access Hash**: When resolving usernames, you must extract and pass the `access_hash` for users and channels.

2. **Message ID Types**: Telegram API uses `int` for IDs, but JSON uses `int64`. Always convert:
   ```go
   ID: int64(msg.ID)
   ```

3. **Response Types**: Telegram API returns variant types (e.g., `MessagesMessagesClass`). Use type switching:
   ```go
   switch m := messagesClass.(type) {
   case *tg.MessagesMessages:
       return m.Messages, m.Users
   case *tg.MessagesMessagesSlice:
       return m.Messages, m.Users
   default:
       return nil, nil
   }
   ```

4. **Peer Types**: Always handle all peer types:
   - `*tg.PeerUser` - Direct messages
   - `*tg.PeerChat` - Legacy groups
   - `*tg.PeerChannel` - Channels and supergroups

---

## Testing

Start the server first, then test commands:

```bash
# Terminal 1: Start server
agent-telegram serve

# Terminal 2: Test commands
agent-telegram ping
agent-telegram get-me
agent-telegram chats -l 20
agent-telegram open @username -l 10
```
