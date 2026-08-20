# JX1 API Gateway

`jx1-api-gateway` is the Go API gateway for GM administration and player-facing account operations in the VLTK/JX1 server stack.

## Responsibilities

- GM authentication with JWTs and role-based authorization.
- Player registration, login, profile, and password changes.
- Account search, account/session inspection, block, unblock, password reset, and kick operations.
- Optional realtime Heaven/S3Relay kick integration.
- Audit log access through MSSQL.
- Embedded, idempotent MSSQL schema migration.
- OpenAPI and Swagger documentation.

The service uses MSSQL as its source of truth. SQLBoiler-generated models are stored in `internal/db/models`; database access is isolated behind repositories, and HTTP requests are adapted by handlers before reaching services.

## Repository layout

```text
cmd/                    Cobra CLI (`web-service start`)
internal/config/         envconfig configuration
internal/http_handler/   Fiber handlers, routes, middleware
internal/service/        admin, player, and auth business logic
internal/repository/     MSSQL repositories
internal/db/models/      SQLBoiler-generated models
pkg/db/                  MSSQL pool and embedded migrations
pkg/heaven/              Heaven/S3Relay protocol client
docs/                    OpenAPI and Swagger assets
```

## API

### System

| Method | Endpoint | Auth |
| --- | --- | --- |
| GET | `/healthz` | Public |
| GET | `/openapi.yaml` | Public |
| GET | `/docs` | Public |

### Authentication

| Method | Endpoint | Auth |
| --- | --- | --- |
| POST | `/api/v1/admin/auth/login` | Public |
| POST | `/api/v1/auth/logout` | Bearer JWT |
| GET | `/api/v1/me` | Bearer JWT |

GM login expects `username` and `password`. The JWT contains the GM role. Supported roles are `viewer`, `operator`, `admin`, and `auditor`.

### Player

| Method | Endpoint | Auth |
| --- | --- | --- |
| POST | `/api/v1/player/accounts/register` | Public |
| POST | `/api/v1/player/auth/login` | Public |
| GET | `/api/v1/player/profile` | Player JWT |
| POST | `/api/v1/player/accounts/change-password` | Player JWT |

### GM administration

All endpoints below require a GM bearer token. Mutation endpoints additionally require the appropriate role.

| Method | Endpoint | Required role |
| --- | --- | --- |
| GET | `/api/v1/admin/accounts?q=&page=&limit=` | GM |
| GET | `/api/v1/admin/accounts/:name` | GM |
| GET | `/api/v1/admin/accounts/:name/sessions` | GM |
| GET | `/api/v1/audit-logs?limit=` | GM |
| POST | `/api/v1/admin/accounts/:name/block` | Operator/Admin |
| POST | `/api/v1/admin/accounts/:name/unblock` | Operator/Admin |
| POST | `/api/v1/admin/accounts/:name/reset-password` | Admin |
| POST | `/api/v1/admin/players/:id/kick` | Operator/Admin |

Mutation requests require a non-empty `reason`; `ticket_id` may be supplied for audit correlation.

## Configuration

Copy the example file and replace placeholders:

```bash
cp .env.example .env
```

Important variables:

```dotenv
PORT=8080
MSSQL_URL=sqlserver://user:password@localhost:1433?database=account_tong&encrypt=disable
JWT_SECRET=replace-with-a-long-random-secret
SESSION_TTL=8h
ENV=development
CORS_ORIGINS=http://localhost:3000
LOG_LEVEL=info
S3RELAY_TARGET=s3relay_ref:5003
HEAVEN_TABLE_PATH=/etc/api-gateway/heaven_table.bin
HEAVEN_SERVER_NAME=replace-me
HEAVEN_SERVER_PASSWORD=replace-me
HEAVEN_IDENTITY=000C-294A-6145
```

Never commit real passwords, JWT secrets, Heaven credentials, or access tokens.

## Running locally

Requirements: Go 1.25+, reachable MSSQL, and the configured account schema.

```bash
cd jx1-api-gateway
cp .env.example .env
go run . web-service start
```

Cobra initializes configuration and runs the embedded idempotent migration before executing the selected command. The server command then opens the MSSQL pool, initializes repositories/services/middleware, and starts Fiber. Stop with `Ctrl-C` for graceful shutdown.

Run tests with:

```bash
go test ./...
```

## Docker

From the parent repository, build or start the gateway using the Compose file that references `./jx1-api-gateway`:

```bash
docker compose -f ../docker-compose.gm.yml up --build api-gateway
```

Mount the Heaven table at the path configured by `HEAVEN_TABLE_PATH`. Set `GM_API_HEAVEN_ACTIONS=1` in the runtime environment to enable the verified Heaven client for realtime kick/block disconnects.

## Documentation

When running, open [http://localhost:8080/docs](http://localhost:8080/docs) for Swagger UI or [http://localhost:8080/openapi.yaml](http://localhost:8080/openapi.yaml) for the raw OpenAPI document.
