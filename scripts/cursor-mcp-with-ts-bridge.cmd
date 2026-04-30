@echo off
setlocal EnableExtensions
REM Cursor MCP entry: TS gRPC bridge (50054) then Go orchestrator on stdio.
REM Do NOT write to stdout except server.exe output (stdio is MCP JSON-RPC).
REM Do NOT use "timeout" here — breaks when stdin is redirected (Cursor MCP).

set "ROOT=%~dp0.."
pushd "%ROOT%\ts-bridge" || exit /b 1
if not exist "dist\index.js" (
  echo [cursor-mcp] ERROR: Missing ts-bridge\dist\index.js. Run npm install and npm run build in ts-bridge. 1>&2
  popd
  exit /b 1
)
where node >nul 2>&1
if errorlevel 1 (
  echo [cursor-mcp] ERROR: node.exe not on PATH 1>&2
  popd
  exit /b 1
)
start "" /B node "dist\index.js"
popd

REM ~2s delay without timeout.exe (stdin-redirect-safe)
ping 127.0.0.1 -n 3 >nul

pushd "%ROOT%\go-orchestrator\bin" || exit /b 1
if not exist "server.exe" (
  echo [cursor-mcp] ERROR: go-orchestrator\bin\server.exe not found 1>&2
  popd
  exit /b 1
)
server.exe %*
set "EXIT=%ERRORLEVEL%"
popd
exit /b %EXIT%
