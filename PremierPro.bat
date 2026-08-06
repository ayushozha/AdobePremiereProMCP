@echo off
REM Double-click this file to launch the PremierPro AI Editor on Windows.

cd /d "%~dp0"

REM Authentication is resolved by the CLI from ANTHROPIC_API_KEY,
REM OPENAI_API_KEY, or %%USERPROFILE%%\.premierpro-mcp\config.json.
REM Claude/Codex OAuth sessions are not API keys and are not scraped here.

for %%T in (node npm npx go) do (
    where %%T >nul 2>&1
    if errorlevel 1 (
        echo   ERROR: %%T is required but was not found on PATH.
        goto :failed
    )
)

REM Ensure CLI dependencies
if not exist "cli\node_modules" (
    echo   Installing CLI dependencies...
    pushd cli
    call npm ci --silent
    if errorlevel 1 goto :subdir_failed
    popd
)

REM Ensure bridge dependencies
if not exist "ts-bridge\node_modules" (
    echo   Installing bridge dependencies...
    pushd ts-bridge
    call npm ci --silent
    if errorlevel 1 goto :subdir_failed
    popd
)

REM A source checkout intentionally does not track generated protobuf stubs.
if not exist "gen\go\premierpro\premiere\v1\premiere.pb.go" (
    where buf >nul 2>&1
    if errorlevel 1 (
        echo   ERROR: buf is required to generate protobuf clients.
        goto :failed
    )
    echo   Generating protobuf clients...
    call buf generate
    if errorlevel 1 goto :failed
)

REM Rebuild through Go's cache so the launcher never runs a stale server.
echo   Building MCP server...
if not exist "go-orchestrator\bin" mkdir "go-orchestrator\bin"
pushd go-orchestrator
call go build -o bin\premierpro-mcp.exe .\cmd\server\
if errorlevel 1 goto :subdir_failed
popd

REM Launch the CLI
call npx --prefix cli tsx cli\src\index.ts
if errorlevel 1 goto :failed
pause
exit /b 0

:subdir_failed
popd

:failed
echo.
echo   PremierPro launcher failed. Fix the error above and try again.
pause
exit /b 1
