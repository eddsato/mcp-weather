# Feature Specification: Open-Meteo Weather MCP Tool

**Status:** Implemented; this specification documents the existing application.

## Summary

Provide an MCP tool that accepts a city or place name, resolves it with the
Open-Meteo Geocoding API, and returns the place's current weather from the
Open-Meteo Forecast API. The server communicates with its MCP client over
standard input and output (stdio).

## User story

As a user of an MCP-compatible AI client, I want to ask for current weather by
place name so that the client can retrieve and present current conditions
without requiring me to know the place's coordinates.

## Functional requirements

1. The server MUST expose a tool named `get_current_weather`.
2. The tool MUST accept a required `location` string containing a city or place
   name.
3. The server MUST trim leading and trailing whitespace and reject an empty
   location with an error.
4. The server MUST resolve the place using the Open-Meteo Geocoding API, asking
   for one result in English.
5. If the geocoding API returns no results, the tool MUST return an error that
   identifies the requested location.
6. The server MUST request current temperature, apparent temperature, relative
   humidity, wind speed, and WMO weather code for the resolved coordinates.
   Temperatures MUST be in Celsius and wind speed MUST be in kilometers per
   hour.
7. On success, the tool MUST return:
   - resolved place name;
   - country when provided by geocoding;
   - current temperature in Celsius;
   - apparent temperature in Celsius;
   - relative humidity percentage;
   - wind speed in kilometers per hour; and
   - WMO weather code.
8. Failures while creating or sending API requests, receiving non-success HTTP
   responses, or decoding API responses MUST be returned as errors with
   contextual information.
9. The server MUST serve MCP protocol messages over stdio. Diagnostic output
   MUST NOT be written to stdout.
10. The server MUST log startup, tool requests, failures, and successful tool
    completion to stderr so logs do not corrupt the MCP protocol stream.
11. When `WEATHER_MCP_LOG_FILE` is set, the server MUST also append those logs
    to the configured file while continuing to write them to stderr.

## Quality requirements

- Outbound HTTP calls MUST have a finite timeout; the current implementation
  uses 10 seconds.
- The MCP request context MUST be propagated to outbound HTTP requests so
  cancellation can stop work.
- Error response bodies MUST be bounded when read.
- The service MUST NOT require an API key for the currently configured
  Open-Meteo endpoints.

## Out of scope

- A REST or HTTP MCP endpoint, AWS API Gateway integration, or deployment
  configuration.
- Forecasts for future dates, historical weather, or multiple locations in one
  call.
- Weather-code-to-description translation in the server.
- User accounts, stored requests, or database persistence.

## Acceptance criteria

- An MCP client can discover `get_current_weather` and invoke it with a
  `location`.
- A valid place returns all specified weather fields with the specified units.
- Empty or unmatched locations return errors rather than successful-looking
  weather data.
- API and JSON errors are surfaced to the MCP client.
- The server runs as an MCP stdio process and does not emit diagnostic text to
  the protocol's stdout stream.
