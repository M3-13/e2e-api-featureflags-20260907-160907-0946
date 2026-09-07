# featureflags

A REST-API for managing and evaluating feature flags, written in Go using only
the standard library (`net/http`). Flags are held in a thread-safe in-memory
store (`sync.RWMutex` + `map`), so the service needs no database or external
dependency to run.

## Tech Stack

- **Language**: Go (1.22+)
- **Web framework**: `net/http` (standard library — no external dependencies)
- **Store**: thread-safe in-memory store with `sync.RWMutex`
- **Tests**: `httptest` + `testing` (standard library)

## Install

Clone the repository and ensure Go 1.22 or newer is installed. No dependencies
need to be downloaded — the project uses only the standard library.

```
go mod download
```

## Run

Start the service with:

```
go run .
```

The server binds to the port from the `PORT` environment variable (default
`8080`):

```
PORT=9090 go run .
```

## Build

```
go build ./...
```

## API

All responses are JSON. Every error response has the shape `{"error": "<message>"}`.

| Method | Path                        | Description                                  |
| ------ | --------------------------- | -------------------------------------------- |
| POST   | `/flags`                    | Create a feature flag (`201` / `400` / `409`) |
| GET    | `/flags`                    | List all feature flags (`200`, empty array when none) |
| GET    | `/flags/{key}`              | Get a single flag (`200` / `404`)            |
| PUT    | `/flags/{key}`              | Update a flag's `enabled`, `description`, `rollout_percent` (`200` / `400` / `404`) |
| DELETE | `/flags/{key}`              | Delete a flag (`204` / `404`)                |
| GET    | `/flags/{key}/evaluate`     | Evaluate a flag for a user (`200` / `400` / `404`) |
| GET    | `/healthz`                  | Health check (`200` `{"status":"ok"}`)       |

### Feature object

```json
{
  "key": "my-flag",
  "enabled": true,
  "description": "optional description",
  "rollout_percent": 50
}
```

## Features

- Create, list, read, update, and delete feature flags.
- Deterministic per-user rollout evaluation.
- Thread-safe in-memory storage.
- Explicit HTTP server timeouts.
- No permissive CORS headers on any endpoint.
