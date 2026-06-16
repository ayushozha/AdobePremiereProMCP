@echo off
REM Install UXP panel for PremierPro MCP transcript bridge on Windows.
REM Requires Premiere Pro 25.0+ and UXP developer mode enabled in Preferences.

set "PLUGIN_NAME=com.premierpro.mcp.uxpbridge"
set "PLUGIN_DIR=%LOCALAPPDATA%\Adobe\UXP\Plugins\External\%PLUGIN_NAME%"
set "SOURCE_DIR=%~dp0..\uxp-panel"

echo Installing UXP transcript bridge panel...
echo Source: %SOURCE_DIR%
echo Target: %PLUGIN_DIR%

if not exist "%SOURCE_DIR%\manifest.json" (
    echo ERROR: uxp-panel source not found at %SOURCE_DIR%
    exit /b 1
)

if exist "%PLUGIN_DIR%" rmdir /s /q "%PLUGIN_DIR%"
mkdir "%PLUGIN_DIR%" 2>nul

xcopy /E /I /Y "%SOURCE_DIR%\*" "%PLUGIN_DIR%\"

echo.
echo UXP panel copied.
echo 1. In Premiere Pro: Preferences ^> UXP Plugins ^> Enable developer mode
echo 2. Restart Premiere Pro
echo 3. Open: Window ^> UXP Plugins ^> PremierPro MCP UXP Bridge
echo 4. Confirm the panel shows Connected to ws://127.0.0.1:9802
echo.
echo Transcript export uses premiere_export_sequence_transcript (not caption tracks).
