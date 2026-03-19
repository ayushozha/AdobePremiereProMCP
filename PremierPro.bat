@echo off
REM Double-click this file to launch the PremierPro AI Editor on Windows.

cd /d "%~dp0"

REM Check for API key
if not defined ANTHROPIC_API_KEY (
    echo.
    echo   ANTHROPIC_API_KEY not found.
    echo.
    echo   Set it with:
    echo     set ANTHROPIC_API_KEY=sk-ant-...
    echo.
    set /p ANTHROPIC_API_KEY="  API Key: "
    echo.
)

REM Ensure CLI dependencies are installed
if not exist "cli\node_modules" (
    echo   Installing dependencies...
    cd cli && call npm install --silent && cd ..
)

REM Ensure MCP server binary exists
if not exist "go-orchestrator\bin\premierpro-mcp.exe" (
    echo   Building MCP server...
    cd go-orchestrator && go build -o bin\premierpro-mcp.exe .\cmd\server\ && cd ..
)

REM Launch the CLI
npx --prefix cli tsx cli\src\index.ts
pause
