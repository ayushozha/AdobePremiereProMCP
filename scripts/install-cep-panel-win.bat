@echo off
REM Install CEP panel for PremierPro MCP on Windows

set PANEL_DIR=%APPDATA%\Adobe\CEP\extensions\com.premierpro.mcp.bridge

echo Installing CEP panel...

REM Remove old installation
if exist "%PANEL_DIR%" rmdir /s /q "%PANEL_DIR%"

REM Create symlink (requires admin on older Windows, works normally on newer)
mklink /D "%PANEL_DIR%" "%~dp0..\cep-panel"

if errorlevel 1 (
    echo Symlink failed. Copying files instead...
    xcopy /E /I /Y "%~dp0..\cep-panel" "%PANEL_DIR%"
)

REM Enable unsigned extensions
REG ADD "HKCU\Software\Adobe\CSXS.11" /v PlayerDebugMode /t REG_SZ /d 1 /f
REG ADD "HKCU\Software\Adobe\CSXS.12" /v PlayerDebugMode /t REG_SZ /d 1 /f
REG ADD "HKCU\Software\Adobe\CSXS.13" /v PlayerDebugMode /t REG_SZ /d 1 /f

echo.
echo CEP panel installed. Restart Premiere Pro to load it.
echo Open: Window - Extensions - PremierPro MCP Bridge
