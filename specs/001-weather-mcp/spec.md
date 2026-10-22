# Feature Specification: Open-Meteo Weather MCP Tool

**Status:** Implemented; this specification documents the existing application.

## Summary

Provide an MCP tool that accepts a city or place name, resolves it with the
Open-Meteo Geocoding API, and returns today's forecast. The tool provides a
follow-up question; after user confirmation, Copilot can invoke the same tool
with the same location to fetch a seven-day forecast. The server communicates
with its MCP client over standard input and output (stdio).

## User story

As a user of an MCP-compatible AI client, I want a tool to give me today's
forecast first, then ask whether I want the seven-day forecast without asking
me to repeat the location.

## Functional requirements

1. The server MUST expose a tool named `get_current_weather`.
2. The tool MUST accept a required `location` string containing a city or place
   name and an optional `include_seven_day_forecast` boolean, which defaults to
   false.
3. The server MUST trim leading and trailing whitespace and reject an empty
   location with an error.
4. The server MUST resolve the place using the Open-Meteo Geocoding API, asking
   for one result in English.
5. If the geocoding API returns no results, the tool MUST return an error that
   identifies the requested location.
6. When `include_seven_day_forecast` is false or omitted, the server MUST
   request current conditions and today's daily forecast only.
7. When `include_seven_day_forecast` is true, the server MUST request current
    conditions and a seven-day daily forecast.
8. Forecast temperatures MUST be in Celsius and wind speed MUST be in
    kilometers per hour.
9. On the initial response, the tool MUST return the resolved location,
    current conditions, today's forecast, and a follow-up question asking
    whether the user wants the seven-day forecast.
10. When the user agrees, Copilot MUST call the tool again with the same
    location and `include_seven_day_forecast` set to true. The tool MUST return
    the seven-day forecast and today's forecast in that response. The tool
    cannot prompt the user directly; Copilot presents the follow-up question.
11. Failures while creating or sending any API request, receiving non-success
    HTTP responses, or decoding API responses MUST be returned as errors with
    contextual information identifying the failed step.
12. The server MUST serve MCP protocol messages over stdio. Diagnostic output
    MUST NOT be written to stdout.
13. The server MUST log startup, tool requests, failures, and successful tool
    completion to stderr so logs do not corrupt the MCP protocol stream.
14. When `WEATHER_MCP_LOG_FILE` is set, the server MUST also append those logs
    to the configured file while continuing to write them to stderr.
15. For each Open-Meteo API call, the server MUST log the HTTP method and URL
    and the raw response body. Since these APIs use GET requests, request
    parameters MUST be represented by the logged URL; no JSON request body is
    sent. Error response bodies MUST remain bounded when read and logged.

## Quality requirements

- Outbound HTTP calls MUST have a finite timeout; the current implementation
  uses 10 seconds.
- The MCP request context MUST be propagated to outbound HTTP requests so
  cancellation can stop work.
- Error response bodies MUST be bounded when read.
- Each tool invocation MUST make no more than two sequential API calls.
- The service MUST NOT require an API key for the configured Open-Meteo APIs.

## Out of scope

- A REST or HTTP MCP endpoint, AWS API Gateway integration, or deployment
  configuration.
- Historical weather or multiple locations in one call.
- Weather-code-to-description translation in the server.
- User accounts, stored requests, or database persistence.

## Acceptance criteria

- An MCP client can discover `get_current_weather` and invoke it with a
  `location`.
- A default invocation returns today's forecast and a follow-up question.
- If the user confirms, a second invocation returns seven forecast days for
  the location from the original interaction.
- Empty or unmatched locations return errors rather than successful-looking
  weather data.
- API and JSON errors are surfaced to the MCP client.
- The server runs as an MCP stdio process and does not emit diagnostic text to
  the protocol's stdout stream.
