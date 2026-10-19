# Open-Meteo Weather MCP Server

A small Go MCP server that provides a `get_current_weather` tool. It resolves a
city or place with the Open-Meteo Geocoding API, then retrieves current weather
from the Open-Meteo Forecast API. No API key is required.

## Requirements

- Go 1.25 or newer

## Run

```sh
go run .
```

The server communicates over stdio, so run it through an MCP client rather than
as a standalone HTTP service.

### Run from PowerShell

From the project directory:

```powershell
Set-Location C:\Users\eduar\mcp
go run .
```

The terminal prints a startup log and then waits for an MCP client. Press
`Ctrl+C` to stop it. When a client invokes the weather tool, request, error, and
completion logs are written to stderr and appear in the terminal.

The stdio protocol requires a client to communicate with the process through
its input and output streams; typing a weather request into the terminal will
not invoke the tool. If VS Code starts the server from `.vscode/mcp.json`, it
starts a separate process. The workspace MCP configuration also sets
`WEATHER_MCP_LOG_FILE`, so the VS Code-launched process writes logs to
`weather-mcp.log` in the project directory as well as stderr. To watch those
logs from a separate PowerShell terminal, restart the MCP server in VS Code,
then run:

```powershell
Get-Content C:\Users\eduar\mcp\weather-mcp.log -Wait -Tail 20
```

Leave that command running and invoke the weather tool from Copilot Chat; each
request and its result or failure will appear in the terminal. Press `Ctrl+C`
to stop watching the file. The same logs are also available in VS Code's Output
panel.

For an MCP client that accepts JSON server configuration, add an entry like:

```json
{
  "mcpServers": {
    "open-meteo-weather": {
      "command": "go",
      "args": ["run", "C:\\path\\to\\weather-mcp"]
    }
  }
}
```

Replace the example path with the absolute path to this project. The client
needs Go installed and network access to the Open-Meteo APIs.

## Tool

`get_current_weather` accepts a `location` string (for example, `"Lisbon"`) and
returns the resolved location, country when available, temperature, apparent
temperature, relative humidity, wind speed, and WMO weather code.
