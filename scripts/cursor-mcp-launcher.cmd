@echo off
rem Opt-in Windows MCP entrypoint. The Go process owns stdin/stdout and its bridge.
"%~dp0..\go-orchestrator\bin\server.exe" --embed-ts-bridge %*
exit /b %errorlevel%
