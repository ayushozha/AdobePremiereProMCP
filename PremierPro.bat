@echo off
REM Double-click this file to launch the PremierPro AI Editor on Windows.

cd /d "%~dp0"

REM Check for API key
if not defined ANTHROPIC_API_KEY (
    REM Try Claude Code auth
    where claude >nul 2>&1
    if %ERRORLEVEL% equ 0 (
        for /f "tokens=*" %%i in ('claude auth print-api-key 2^>nul') do set "ANTHROPIC_API_KEY=%%i"
    )
)

REM If still no key, offer to login via Claude Code
if not defined ANTHROPIC_API_KEY (
    where claude >nul 2>&1
    if %ERRORLEVEL% equ 0 (
        echo.
        echo   No API key found. Launching Claude login...
        echo.
        call claude login
        for /f "tokens=*" %%i in ('claude auth print-api-key 2^>nul') do set "ANTHROPIC_API_KEY=%%i"
    )
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
