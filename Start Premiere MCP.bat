@echo off
REM Convenience launcher: opens Premiere Pro with a project and starts the
REM backend services in one step. You do not need it for normal use --
REM quitting Premiere normally and reopening a project works on its own.
REM Its one real job is recovering after a crash or force-kill, which can
REM leave orphaned CEPHtmlEngine processes that break extension loading.
REM
REM If Premiere is already open this stops and tells you to quit it first,
REM because extensions register at startup. Add -Force to close it for you
REM (save your work first). See scripts\launch-premiere-mcp.ps1 for detail.

powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\launch-premiere-mcp.ps1" %*
pause
