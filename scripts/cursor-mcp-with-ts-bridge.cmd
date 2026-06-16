@echo off
setlocal EnableExtensions
REM Delegates to Node so ts-bridge is a child process (not start /B). When MCP stops,
REM launcher tears down bridge + server — Premiere CEP disconnects instead of a stale WS.

where node >nul 2>&1 || (
  echo [cursor-mcp] ERROR: node.exe not on PATH 1>&2
  exit /b 1
)
node "%~dp0cursor-mcp-launcher.cjs" %*
exit /b %ERRORLEVEL%
