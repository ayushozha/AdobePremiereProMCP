#!/bin/bash
# Install the CEP panel into Adobe Premiere Pro's extensions folder

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

PANEL_DIR="$HOME/Library/Application Support/Adobe/CEP/extensions/com.premierpro.mcp.bridge"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}Installing CEP panel for PremierPro MCP...${NC}"
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
defaults write com.adobe.CSXS.11 PlayerDebugMode 1
echo -e "${GREEN}Enabled unsigned extensions (CSXS.11 PlayerDebugMode=1)${NC}"
echo ""

echo -e "${GREEN}CEP panel installed. Restart Premiere Pro to load it.${NC}"
