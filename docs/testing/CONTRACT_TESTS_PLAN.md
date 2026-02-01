# Contract Tests Plan

План внедрения contract-тестов с fixtures из реальных ответов Telegram API.

## Обзор

Contract-тесты проверяют, что наш код корректно обрабатывает реальные ответы Telegram API. Вместо моков с выдуманными данными используем записанные fixtures — "снимки" реальных ответов.

### Преимущества

- Тесты работают с реальной структурой данных Telegram
- Ловят breaking changes в API
- Не требуют сети при запуске
- Быстрые и детерминированные
- Можно запускать в CI

### Ограничения

- Fixtures могут устареть (Telegram меняет API)
- Нужно периодически обновлять fixtures
- Первоначальная запись требует реальный аккаунт

---

## Архитектура

```
testdata/
├── fixtures/
│   ├── messages/
│   │   ├── get_history_private_chat.json
│   │   ├── get_history_group.json
│   │   ├── get_history_channel.json
│   │   ├── send_message_success.json
│   │   ├── send_message_flood_wait.json
│   │   └── get_dialogs.json
│   ├── contacts/
│   │   ├── resolve_username_user.json
│   │   ├── resolve_username_channel.json
│   │   └── resolve_username_not_found.json
│   ├── channels/
│   │   ├── get_participants.json
│   │   ├── get_full_channel.json
│   │   └── create_channel.json
│   ├── users/
│   │   ├── get_full_user.json
│   │   └── get_full_user_bot.json
│   └── errors/
│       ├── flood_wait_420.json
│       ├── peer_id_invalid.json
│       ├── chat_write_forbidden.json
│       └── user_banned_in_channel.json
├── recorder/
│   └── main.go              # Утилита записи fixtures
└── README.md
```

---

## Приоритеты записи fixtures

### P0 — Критические (записать первыми)

| Метод | Fixture | Описание |
|-------|---------|----------|
| `MessagesGetHistory` | `get_history_*.json` | Получение сообщений — основная функция |
| `MessagesGetDialogs` | `get_dialogs.json` | Список чатов |
| `ContactsResolveUsername` | `resolve_username_*.json` | Резолв peer — используется везде |
| `MessagesSendMessage` | `send_message_*.json` | Отправка сообщений |

### P1 — Важные

| Метод | Fixture | Описание |
|-------|---------|----------|
| `ChannelsGetParticipants` | `get_participants.json` | Список участников |
| `UsersGetFullUser` | `get_full_user.json` | Информация о пользователе |
| `MessagesGetFullChat` | `get_full_chat.json` | Информация о чате |
| `ChannelsGetFullChannel` | `get_full_channel.json` | Информация о канале |
| `MessagesSearch` | `search_messages.json` | Поиск сообщений |
| `ContactsSearch` | `search_contacts.json` | Глобальный поиск |

### P2 — Второстепенные

| Метод | Fixture | Описание |
|-------|---------|----------|
| `MessagesForwardMessages` | `forward_messages.json` | Пересылка |
| `MessagesEditMessage` | `edit_message.json` | Редактирование |
| `MessagesSendMedia` | `send_photo.json`, `send_file.json` | Медиа |
| `MessagesGetForumTopics` | `forum_topics.json` | Темы форума |

### P3 — Error cases

| Ошибка | Fixture | Когда возникает |
|--------|---------|-----------------|
| `FLOOD_WAIT_X` | `flood_wait_420.json` | Rate limit |
| `PEER_ID_INVALID` | `peer_id_invalid.json` | Неверный peer |
| `CHAT_WRITE_FORBIDDEN` | `chat_write_forbidden.json` | Нет прав писать |
| `USER_BANNED_IN_CHANNEL` | `user_banned.json` | Пользователь забанен |
| `USERNAME_NOT_OCCUPIED` | `username_not_found.json` | Username не существует |

---

## Формат fixtures

### Структура файла

```json
{
  "meta": {
    "method": "messages.getHistory",
    "recorded_at": "2024-01-15T10:30:00Z",
    "telegram_layer": 185,
    "notes": "Private chat with regular user"
  },
  "request": {
    "peer": {
      "_": "inputPeerUser",
      "user_id": 123456789,
      "access_hash": 1234567890123456789
    },
    "limit": 50,
    "offset_id": 0
  },
  "response": {
    "_": "messages.messages",
    "messages": [...],
    "chats": [...],
    "users": [...]
  }
}
```

### Поля meta

| Поле | Описание |
|------|----------|
| `method` | Название метода Telegram API |
| `recorded_at` | Дата записи (ISO 8601) |
| `telegram_layer` | Версия layer API |
| `notes` | Контекст: тип чата, особенности |
| `sanitized` | Были ли удалены личные данные |

---

## Sanitization (очистка данных)

Перед коммитом fixtures нужно очистить от личных данных:

### Что заменять

| Поле | Замена |
|------|--------|
| `user_id` | `100000001`, `100000002`, ... |
| `chat_id` | `200000001`, `200000002`, ... |
| `channel_id` | `300000001`, `300000002`, ... |
| `access_hash` | `1000000000000000001`, ... |
| `phone` | `+15551234567` |
| `first_name` | `Test`, `User`, `Alice`, `Bob` |
| `last_name` | `User`, `Smith`, `Jones` |
| `username` | `testuser1`, `testuser2`, ... |
| `message` (text) | `Test message 1`, `Hello world`, ... |

### Что сохранять

- Структуру ответа
- Типы полей (`_` discriminator)
- Флаги и настройки
- Timestamps (можно сдвинуть)
- Размеры, counts, offsets

---

## Инструменты

### 1. Recorder — запись fixtures

```go
// testdata/recorder/main.go
// Запускается вручную с реальным аккаунтом
// Записывает ответы API в JSON файлы

go run ./testdata/recorder -method messages.getHistory -peer @username -output fixtures/messages/
```

### 2. Sanitizer — очистка данных

```go
// testdata/sanitizer/main.go
// Заменяет личные данные на тестовые

go run ./testdata/sanitizer -input fixtures/raw/ -output fixtures/
```

### 3. Validator — проверка fixtures

```go
// testdata/validator/main.go
// Проверяет, что fixtures соответствуют текущим типам gotd/td

go run ./testdata/validator ./fixtures/...
```

---

## Использование в тестах

### Загрузка fixture

```go
// telegram/message/messages_test.go

func TestGetMessages_ParsesHistory(t *testing.T) {
    fixture := testutil.LoadFixture(t, "messages/get_history_private_chat.json")

    // Десериализуем response в тип gotd/td
    var resp tg.MessagesMessages
    require.NoError(t, json.Unmarshal(fixture.Response, &resp))

    // Тестируем нашу логику обработки
    result, err := convertMessages(&resp)
    require.NoError(t, err)

    assert.Len(t, result.Messages, 50)
    assert.Equal(t, "Test message 1", result.Messages[0].Text)
}
```

### Table-driven tests с fixtures

```go
func TestResolveUsername(t *testing.T) {
    tests := []struct {
        name    string
        fixture string
        wantErr bool
    }{
        {"user", "contacts/resolve_username_user.json", false},
        {"channel", "contacts/resolve_username_channel.json", false},
        {"not_found", "contacts/resolve_username_not_found.json", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            fixture := testutil.LoadFixture(t, tt.fixture)
            // ...
        })
    }
}
```

---

## CI/CD интеграция

### Makefile targets

```makefile
.PHONY: test-contracts
test-contracts:
	go test -v -tags=contracts ./...

.PHONY: validate-fixtures
validate-fixtures:
	go run ./testdata/validator ./testdata/fixtures/...

.PHONY: update-fixtures
update-fixtures:
	@echo "Run manually with authenticated session"
	@echo "go run ./testdata/recorder ..."
```

### GitHub Actions

```yaml
# .github/workflows/test.yml
- name: Validate fixtures
  run: make validate-fixtures

- name: Run contract tests
  run: make test-contracts
```

---

## Roadmap

### Фаза 1: Инфраструктура (1-2 дня)

- [ ] Создать структуру `testdata/fixtures/`
- [ ] Написать `testutil.LoadFixture()` helper
- [ ] Создать базовый recorder

### Фаза 2: P0 Fixtures (2-3 дня)

- [ ] Записать `get_history` для разных типов чатов
- [ ] Записать `get_dialogs`
- [ ] Записать `resolve_username` варианты
- [ ] Записать `send_message`
- [ ] Sanitize все fixtures

### Фаза 3: Тесты (3-5 дней)

- [ ] Тесты для `telegram/message/messages.go`
- [ ] Тесты для `telegram/message/convert.go`
- [ ] Тесты для `telegram/peer.go`
- [ ] Тесты для `telegram/chat/dialogs.go`

### Фаза 4: P1 Fixtures + Error cases (2-3 дня)

- [ ] Записать P1 fixtures
- [ ] Записать error responses
- [ ] Тесты error handling

### Фаза 5: CI интеграция (1 день)

- [ ] Validator в CI
- [ ] Contract tests в CI
- [ ] Документация

---

## Риски и митигация

| Риск | Митигация |
|------|-----------|
| Fixtures устаревают | Validator проверяет совместимость с текущими типами |
| Telegram меняет API | Периодический ре-рекординг (раз в 3-6 месяцев) |
| Утечка личных данных | Обязательный sanitization перед коммитом |
| Большой размер fixtures | Минимизировать данные, хранить только нужные поля |

---

## Ссылки

- [gotd/td типы](https://pkg.go.dev/github.com/gotd/td/tg)
- [Telegram API документация](https://core.telegram.org/api)
- [Telegram Layer версии](https://core.telegram.org/api/layers)
