@echo off
REM Double-click this file to launch the PremierPro AI Editor on Windows.

cd /d "%~dp0"

REM Check for existing API keys
if defined ANTHROPIC_API_KEY goto :has_auth
if defined OPENAI_API_KEY goto :has_auth

REM Try Claude Code auth
where claude >nul 2>&1
if %ERRORLEVEL% equ 0 (
    for /f "tokens=*" %%i in ('claude auth print-api-key 2^>nul') do set "ANTHROPIC_API_KEY=%%i"
    if defined ANTHROPIC_API_KEY goto :has_auth
)

REM Try Codex auth
where codex >nul 2>&1
if %ERRORLEVEL% equ 0 (
    for /f "tokens=*" %%i in ('codex auth print-api-key 2^>nul') do set "OPENAI_API_KEY=%%i"
    if defined OPENAI_API_KEY goto :has_auth
)

REM No auth found — prompt
echo.
echo   No API key found. Choose a provider to login:
echo     1) Claude (Anthropic)
echo     2) OpenAI / Codex
echo.
set /p choice="  Choice [1]: "

if "%choice%"=="2" (
    where codex >nul 2>&1
    if %ERRORLEVEL% equ 0 (
        call codex login
        for /f "tokens=*" %%i in ('codex auth print-api-key 2^>nul') do set "OPENAI_API_KEY=%%i"
    ) else (
        set /p OPENAI_API_KEY="  Paste your OpenAI API key: "
    )
) else (
    where claude >nul 2>&1
    if %ERRORLEVEL% equ 0 (
        call claude login
        for /f "tokens=*" %%i in ('claude auth print-api-key 2^>nul') do set "ANTHROPIC_API_KEY=%%i"
    ) else (
        set /p ANTHROPIC_API_KEY="  Paste your Anthropic API key: "
    )
)

:has_auth

REM Ensure CLI dependencies
if not exist "cli\node_modules" (
    echo   Installing CLI dependencies...
    cd cli && call npm install --silent && cd ..
)

REM Ensure bridge dependencies
if not exist "ts-bridge\node_modules" (
    echo   Installing bridge dependencies...
    cd ts-bridge && call npm install --silent && cd ..
)

REM Ensure MCP server binary
if not exist "go-orchestrator\bin\premierpro-mcp.exe" (
    echo   Building MCP server...
    cd go-orchestrator && go build -o bin\premierpro-mcp.exe .\cmd\server\ && cd ..
)

REM Launch the CLI
npx --prefix cli tsx cli\src\index.ts
pause
