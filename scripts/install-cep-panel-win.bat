@echo off
REM Install CEP panel for PremierPro MCP on Windows

set PANEL_DIR=%APPDATA%\Adobe\CEP\extensions\com.premierpro.mcp.bridge

echo Installing CEP panel...

REM Remove old installation
if exist "%PANEL_DIR%" rmdir /s /q "%PANEL_DIR%"

REM Copy files (not a symlink): Adobe CEP's signature verification fails on
REM symlinked extension folders unless the host runs as admin - see
REM https://github.com/Adobe-CEP/CEP-Resources/blob/master/ZXPSignCMD/KnownIssue2024.md
xcopy /E /I /Y "%~dp0..\cep-panel" "%PANEL_DIR%"

REM Enable unsigned extensions
REG ADD "HKCU\Software\Adobe\CSXS.11" /v PlayerDebugMode /t REG_SZ /d 1 /f
REG ADD "HKCU\Software\Adobe\CSXS.12" /v PlayerDebugMode /t REG_SZ /d 1 /f
REG ADD "HKCU\Software\Adobe\CSXS.13" /v PlayerDebugMode /t REG_SZ /d 1 /f

echo.
echo CEP panel installed. Restart Premiere Pro to load it.
echo Open: Window - Extensions - PremierPro MCP Bridge
