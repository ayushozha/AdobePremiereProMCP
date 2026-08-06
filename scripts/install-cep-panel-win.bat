@echo off
REM Install CEP panel for PremierPro MCP on Windows

set "CEP_EXTENSIONS_DIR=%APPDATA%\Adobe\CEP\extensions"
set "PANEL_DIR=%CEP_EXTENSIONS_DIR%\com.premierpro.mcp.bridge"
set "PANEL_SOURCE=%~dp0..\cep-panel\dist"

echo Installing CEP panel...

REM Build a production extension. The source .debug file is development-only.
call :build_panel
if errorlevel 1 goto :failed

if not exist "%CEP_EXTENSIONS_DIR%" mkdir "%CEP_EXTENSIONS_DIR%"
if errorlevel 1 goto :failed

REM Remove old installation
if exist "%PANEL_DIR%" rmdir /s /q "%PANEL_DIR%"
if exist "%PANEL_DIR%" goto :failed

REM Create symlink (requires admin on older Windows, works normally on newer)
mklink /D "%PANEL_DIR%" "%PANEL_SOURCE%"

if errorlevel 1 (
    echo Symlink failed. Copying files instead...
    xcopy /E /I /Y "%PANEL_SOURCE%" "%PANEL_DIR%" >nul
    if errorlevel 1 goto :failed
)

REM Enable unsigned extensions
REG ADD "HKCU\Software\Adobe\CSXS.11" /v PlayerDebugMode /t REG_SZ /d 1 /f
if errorlevel 1 goto :failed
REG ADD "HKCU\Software\Adobe\CSXS.12" /v PlayerDebugMode /t REG_SZ /d 1 /f
if errorlevel 1 goto :failed
REG ADD "HKCU\Software\Adobe\CSXS.13" /v PlayerDebugMode /t REG_SZ /d 1 /f
if errorlevel 1 goto :failed

echo.
echo CEP panel installed. Restart Premiere Pro to load it.
echo Open: Window - Extensions - PremierPro MCP Bridge
exit /b 0

:build_panel
pushd "%~dp0..\cep-panel" || exit /b 1
if not exist "node_modules\ws" call npm ci --omit=dev
if errorlevel 1 (
    popd
    exit /b 1
)
call npm run build
set "BUILD_RESULT=%ERRORLEVEL%"
popd
exit /b %BUILD_RESULT%

:failed
echo.
echo CEP panel installation failed. Review the error above; no success is assumed.
exit /b 1
