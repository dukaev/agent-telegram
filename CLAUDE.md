# agent-telegram

Go CLI + IPC server for Telegram, distributed via npm. Uses gotd/td (MTProto).

## Commands

```bash
make build              # Build binary
make test               # Run tests
make lint               # Lint (golangci-lint + revive)
make dev                # Live reload (air) for serve
make release            # Patch release (v0.4.1 → v0.4.2)
make release-minor      # Minor release (v0.4.1 → v0.5.0)
make release-major      # Major release (v0.4.1 → v1.0.0)
make release-local      # Test release locally (goreleaser snapshot)
go build ./...          # Verify compilation
go vet ./...            # Verify no warnings
```

## Release flow

1. Commit changes to `main`
2. Run `make release` (or `release-minor` / `release-major`)
3. GitHub Actions handles the rest: GoReleaser builds binaries, creates GitHub release, publishes to npm
4. Version in `cmd/root.go` stays `"dev"` — GoReleaser injects version via `-ldflags -X agent-telegram/cmd.version=VERSION`
5. For local builds with version: `go build -ldflags "-X agent-telegram/cmd.version=v0.5.0" -o binaries/agent-telegram-darwin-arm64 .`

## Architecture

```
cmd/                          CLI commands (cobra)
  ├── root.go                 Root command, groups, global flags
  ├── register.go             Command registration (imports + init)
  └── <domain>/cmd.go         Parent command + AddXxxCommand()

telegram/                     Telegram client layer
  ├── client.go               Main client (init, start, domain wiring)
  ├── accessors.go            Domain client accessors (Chat(), Gift(), etc.)
  ├── domain_interfaces.go    Interfaces for all domain clients
  ├── types/                  Shared types (Params, Result, entities)
  │   └── types_<domain>.go   Per-domain types
  └── <domain>/client.go      Domain client (embeds BaseClient)

internal/
  ├── cliutil/                CLI utilities (Runner, Recipient, Pagination, print helpers)
  └── telegram/ipc/
      ├── getme.go            IPC Client interface
      ├── handlers.go         Handler functions
      ├── handler.go          Generic Handler[T,R] factory
      └── register.go         Method → handler map
```

## Adding a new domain

1. **Types**: `telegram/types/types_<domain>.go` — Params (with `Validate() error`), Result, entity structs
2. **Client**: `telegram/<domain>/client.go` — embed `*client.BaseClient`, implement methods
3. **Interface**: `telegram/domain_interfaces.go` — add `<Domain>Client` interface
4. **Wiring**: `telegram/client.go` — add field, import, `initDomainClients()`, `setDomainAPIs()`
5. **Accessor**: `telegram/accessors.go` — add `<Domain>() <Domain>Client`
6. **IPC interface**: `internal/telegram/ipc/getme.go` — add to `Client` interface
7. **IPC handlers**: `internal/telegram/ipc/handlers.go` — handler functions using `Handler()`
8. **IPC register**: `internal/telegram/ipc/register.go` — add to `methodHandlers` map
9. **CLI**: `cmd/<domain>/cmd.go` + subcommand files
10. **Register CLI**: `cmd/root.go` (blank import) + `cmd/register.go` (import + `Add...Command`)

## Patterns

- Domain client methods: `func (c *Client) Method(ctx, params Type) (*ResultType, error)` — always start with `c.CheckInitialized()`
- Peer resolution: `c.ResolvePeer(ctx, params.Peer)` returns `tg.InputPeerClass`
- Params validation: use `validate:"required"` tags + `ValidateStruct(p)`, or custom `Validate()` for complex logic
- IPC handlers are one-liners: `func handler(c Client) HandlerFunc { return Handler(c.Domain().Method, "name") }`
- CLI commands use `cliutil.NewRunnerFromCmd()` → `runner.CallWithParams()` → `runner.PrintResult()`
- Recipient flag: `cliutil.Recipient` type with `--to` / `-t`, call `to.AddToParams(params)`
- Pagination: `cliutil.NewPagination(limit, offset, cfg)` → `pag.ToParams(params, includeOffset)`
- Success output: `cliutil.PrintSuccessSummary(result, "message")`

## Conventions

- Go module: `agent-telegram`
- Telegram API via gotd/td v0.137.0 (`github.com/gotd/td/tg`)
- CLI framework: cobra (`github.com/spf13/cobra`)
- Status/logs to stderr, data to stdout
- JSON tags: camelCase (`json:"fieldName"`)
- Command groups: `auth`, `message`, `chat`, `server`
- No hardcoded version in code — use ldflags
