@echo off
REM Start the PremierPro MCP backend services on Windows.
REM
REM Open the CEP panel in Premiere FIRST (Window > Extensions > PremierPro MCP
REM Bridge). The panel hosts the WebSocket server on port 9801 that ts-bridge
REM connects to as a client; if the panel is not up yet, ts-bridge starts in
REM disconnected mode and you have to restart it.
REM
REM The toolchain on this machine lives outside the default install locations
REM (portable builds), so PATH is set explicitly here rather than relying on
REM whatever the calling shell happens to have.

setlocal

set "PATH=%LOCALAPPDATA%\Programs\Go\bin;%PATH%"
set "PATH=%LOCALAPPDATA%\Programs\Python312;%LOCALAPPDATA%\Programs\Python312\Scripts;%PATH%"
set "PATH=%LOCALAPPDATA%\Microsoft\WinGet\Links;%PATH%"
set "PATH=%LOCALAPPDATA%\Microsoft\WinGet\Packages\OpenJS.NodeJS.LTS_Microsoft.Winget.Source_8wekyb3d8bbwe\node-v24.18.0-win-x64;%PATH%"
set "PATH=%USERPROFILE%\.cargo\bin;%PATH%"
set "PATH=%LOCALAPPDATA%\Microsoft\WinGet\Packages\Google.Protobuf_Microsoft.Winget.Source_8wekyb3d8bbwe\bin;%PATH%"
REM `just` shells out to sh, which ships with Git for Windows.
set "PATH=C:\Program Files\Git\usr\bin;%PATH%"

cd /d "%~dp0.."

echo Starting PremierPro MCP services (rust :50052, python :50053, ts-bridge :50054)...
bash scripts/start-all.sh

endlocal
