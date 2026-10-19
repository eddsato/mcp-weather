# Implementation Tasks: Open-Meteo Weather MCP Tool

This task list was reconstructed after implementation. Completed items are
marked `[x]`; unchecked items are suggested verification work, not claims about
work already performed.

## Build

- [x] Create the Go module and add the MCP Go SDK dependency.
- [x] Define typed MCP input/output and Open-Meteo response structures.
- [x] Implement geocoding and current-forecast requests.
- [x] Add location validation, HTTP timeout, context cancellation, response
  checks, bounded error-body reading, and JSON decoding errors.
- [x] Register `get_current_weather` and serve it over MCP stdio.
- [x] Add the VS Code MCP launch configuration.
- [x] Format the Go source and verify the module builds with `go build ./...`.

## Follow-up verification

- [ ] Add unit tests for blank input, no geocoding results, successful response
  mapping, and provider/decoding failures.
- [ ] Add an MCP-client smoke test that verifies tool discovery and invocation
  without relying on a manually started stdio process.
