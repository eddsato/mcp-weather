# Project Constitution

## Purpose

This project provides a small weather lookup tool to MCP-compatible AI clients.
Keep the implementation focused on resolving a place and returning its current
weather from Open-Meteo.

## Principles

1. **MCP-first interface** — The supported client interface is an MCP tool over
   stdio. Do not treat the server as an HTTP service unless a separate feature
   is specified.
2. **Small, typed Go implementation** — Use the Go standard library and the
   existing MCP Go SDK where practical. Keep API request and response data
   explicitly typed.
3. **Explicit failures** — Return actionable errors for invalid input, missing
   locations, HTTP failures, and invalid API responses. Do not silently replace
   failed data with plausible values.
4. **Protect the protocol stream** — Reserve stdout for MCP stdio messages. Send
   diagnostic logs to stderr.
5. **Bound external calls** — Apply timeouts to outbound API requests and
   propagate cancellation from the MCP request context.
6. **No unnecessary secrets or persistence** — The Open-Meteo API currently
   requires no API key, and this service does not need to store user data.
   Reassess both assumptions if a future provider or feature changes them.

## Change expectations

- Keep the user-visible tool schema and this specification in sync.
- Add or update focused tests when behavior changes.
- Run `gofmt` and `go build ./...` after Go code changes; run tests when present.
- Keep deployment, authentication, and HTTP transport concerns out of scope
  unless separately specified.
