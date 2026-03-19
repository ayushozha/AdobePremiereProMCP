#!/bin/bash
# Install the CEP panel into Adobe Premiere Pro's extensions folder
# Supports macOS and Linux

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RED='\033[0;31m'
NC='\033[0m'

# Detect OS and set the appropriate extensions directory
OS="$(uname -s)"
case "$OS" in
    Darwin)
        PANEL_DIR="$HOME/Library/Application Support/Adobe/CEP/extensions/com.premierpro.mcp.bridge"
        ;;
    Linux)
        PANEL_DIR="$HOME/.local/share/Adobe/CEP/extensions/com.premierpro.mcp.bridge"
        ;;
    MINGW*|MSYS*|CYGWIN*)
        echo -e "${RED}Windows detected. Please use install-cep-panel-win.bat instead.${NC}"
        exit 1
        ;;
    *)
        echo -e "${RED}Unsupported OS: $OS${NC}"
        exit 1
        ;;
esac

echo -e "${CYAN}Installing CEP panel for PremierPro MCP...${NC}"
echo -e "${CYAN}Detected OS: $OS${NC}"
echo ""

# Ensure the parent directory exists
mkdir -p "$(dirname "$PANEL_DIR")"

# Remove old installation
if [ -e "$PANEL_DIR" ] || [ -L "$PANEL_DIR" ]; then
    echo -e "${YELLOW}Removing old installation...${NC}"
    rm -rf "$PANEL_DIR"
fi

# Symlink to our cep-panel directory
ln -s "$PROJECT_ROOT/cep-panel" "$PANEL_DIR"
echo -e "${GREEN}Symlinked:${NC}"
echo "  $PROJECT_ROOT/cep-panel"
echo "  -> $PANEL_DIR"
echo ""

# Enable unsigned extensions for debugging
case "$OS" in
    Darwin)
        defaults write com.adobe.CSXS.11 PlayerDebugMode 1
        defaults write com.adobe.CSXS.12 PlayerDebugMode 1
        defaults write com.adobe.CSXS.13 PlayerDebugMode 1
        echo -e "${GREEN}Enabled unsigned extensions (CSXS.11-13 PlayerDebugMode=1)${NC}"
        ;;
    Linux)
        # On Linux, CEP debug mode is set via a file
        CSXS_PREFS_DIR="$HOME/.adobe"
        mkdir -p "$CSXS_PREFS_DIR"
        for ver in 11 12 13; do
            PREFS_FILE="$CSXS_PREFS_DIR/CSXS.${ver}.prefs"
            if [ -f "$PREFS_FILE" ]; then
                # Remove existing PlayerDebugMode line if present
                sed -i '/PlayerDebugMode/d' "$PREFS_FILE"
            fi
            echo "PlayerDebugMode=1" >> "$PREFS_FILE"
        done
        echo -e "${GREEN}Enabled unsigned extensions (CSXS.11-13 PlayerDebugMode=1)${NC}"
        ;;
esac
echo ""

echo -e "${GREEN}CEP panel installed. Restart Premiere Pro to load it.${NC}"
echo "Open: Window > Extensions > PremierPro MCP Bridge"
