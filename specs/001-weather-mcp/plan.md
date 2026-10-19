# Implementation Plan: Open-Meteo Weather MCP Tool

**Status:** Implemented; this plan records the existing design.

## Technical context

- **Language:** Go 1.25 or newer.
- **MCP SDK:** `github.com/modelcontextprotocol/go-sdk` v1.7.0.
- **Transport:** MCP stdio.
- **Weather provider:** Open-Meteo Geocoding and Forecast APIs; no API key is
  currently required.
- **Persistence:** None.

## Design

1. Define typed Go structures for the MCP tool input/output and for the subset
   of each Open-Meteo response used by the application.
2. Register `get_current_weather` with the MCP server. Derive its input schema
   from the typed input structure.
3. In the tool handler:
   - trim and validate the location;
   - URL-encode and send a geocoding query;
   - report an error if no place is found;
   - use the first returned place's coordinates to request current conditions;
   - map the provider response into the public tool output.
4. Use a shared `http.Client` with a 10-second timeout. Attach the MCP context
   to outbound requests, check HTTP status, limit error-body reads, and decode
   JSON into typed values.
5. Run the MCP server with the SDK's stdio transport. Use stderr for any
   diagnostics so stdout remains reserved for protocol messages.
6. Configure VS Code to launch the Go module as the MCP server.

## Current implementation mapping

| Concern | Existing implementation |
| --- | --- |
| MCP tool and server | `main.go` |
| Tool input/output and provider response types | `main.go` |
| Geocoding and forecast requests | `main.go` |
| Module and SDK dependencies | `go.mod`, `go.sum` |
| VS Code stdio launch configuration | `.vscode/mcp.json` |

## Verification plan

- Format Go source with `gofmt`.
- Build the module with `go build ./...`.
- Run the server through an MCP client and invoke `get_current_weather` with a
  known place.
- Add automated tests for input validation, provider errors, JSON decoding, and
  response mapping as the implementation is extended.

The source has been formatted and `go build ./...` succeeded during setup.
Automated tests are a follow-up; no test files are currently part of the
project.
