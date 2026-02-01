# Test Data

This directory contains fixtures for contract testing against Telegram API responses.

## Structure

```
testdata/
├── fixtures/           # Recorded API responses (sanitized)
│   ├── messages/       # Messages.* methods
│   ├── contacts/       # Contacts.* methods
│   ├── channels/       # Channels.* methods
│   ├── users/          # Users.* methods
│   └── errors/         # Error responses
├── recorder/           # Tool to record new fixtures
└── README.md
```

## Usage in Tests

```go
import "agent-telegram/internal/testutil"

func TestSomething(t *testing.T) {
    // Load a fixture
    fixture := testutil.LoadFixture(t, "messages/get_history_private_chat.json")

    // Use the response
    var resp tg.MessagesMessages
    json.Unmarshal(fixture.Response, &resp)

    // Test your code
    result := processMessages(&resp)
    assert.Equal(t, 3, len(result))
}
```

## Recording New Fixtures

1. Ensure you have an authenticated session
2. Run the recorder:

```bash
go run ./testdata/recorder \
    -method messages.getHistory \
    -peer @username \
    -notes "private_chat" \
    -output ./testdata/fixtures
```

3. Sanitize personal data before committing (see below)

## Sanitization Rules

Before committing fixtures, replace personal data:

| Original | Replacement |
|----------|-------------|
| Real user IDs | `100000001`, `100000002`, ... |
| Real chat IDs | `200000001`, `200000002`, ... |
| Real channel IDs | `300000001`, `300000002`, ... |
| Access hashes | `1000000000000000001`, ... |
| Phone numbers | `+15551234567` |
| Real names | `Alice`, `Bob`, `Test`, `User` |
| Real usernames | `testuser1`, `testuser2`, ... |
| Message content | `Test message 1`, `Hello world`, ... |

## Fixture Format

```json
{
  "meta": {
    "method": "messages.getHistory",
    "recorded_at": "2024-01-15T10:30:00Z",
    "telegram_layer": 185,
    "notes": "Description of this fixture",
    "sanitized": true
  },
  "request": { ... },
  "response": { ... }
}
```

For error fixtures, include an `error` field:

```json
{
  "meta": { ... },
  "request": { ... },
  "response": null,
  "error": {
    "code": 400,
    "type": "USERNAME_NOT_OCCUPIED",
    "message": "The username is not in use by anyone"
  }
}
```

## Updating Fixtures

Fixtures may become outdated when Telegram updates their API. To refresh:

1. Check the current Telegram layer version
2. Re-record affected fixtures
3. Update `telegram_layer` in meta
4. Sanitize and commit

## Priority List

### P0 - Critical (must have)
- [ ] `messages/get_history_private_chat.json`
- [ ] `messages/get_history_group.json`
- [ ] `messages/get_dialogs.json`
- [ ] `contacts/resolve_username_user.json`
- [ ] `contacts/resolve_username_channel.json`

### P1 - Important
- [ ] `channels/get_participants.json`
- [ ] `users/get_full_user.json`
- [ ] `messages/search.json`

### P2 - Nice to have
- [ ] `messages/send_message.json`
- [ ] `channels/get_full_channel.json`
- [ ] Various error cases
