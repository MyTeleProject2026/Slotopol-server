# TiDB / MySQL operational logs

Slotopol-server uses the existing XORM SQL storage for operational logs. TiDB is MySQL-compatible, so the `operational_logs` model is synchronized by the server at runtime; no manual SQL is required for deployment.

## Table

The server creates/updates `operational_logs` from the Go model in `api/operations.go` when the operations API is first used. The table contains indexed fields for timestamp, level, event type, request ID, user ID, club ID, game ID, endpoint and status, plus diagnostic message/error/metadata fields.

## Admin API

All three endpoints are protected by the existing `ALadmin` RBAC permission:

- `GET /admin/operations/summary`
- `GET /admin/operations/logs?limit=100&level=ERROR`
- `GET /admin/operations/stream`

The stream uses Server-Sent Events (SSE) and does not put access tokens in URLs.

## Production database requirement

Configure Slotopol-server with the existing TiDB MySQL connection (`SLOTOPOL_DBDRIVER=mysql` and the existing club/spin DSNs). Do not create a second logging database. Do not run SQL manually for the operational log table.

The application must have permission to create/alter the `operational_logs` table during its normal startup/first-use migration path. If the TiDB account is intentionally restricted from schema changes, the deployment process should grant that migration permission temporarily or use the project's migration mechanism before restricting it.

## Security

Operational logging deliberately excludes request bodies, passwords, Authorization headers and JWT values. The admin stream is available only after server-side authentication and `ALadmin` authorization.
