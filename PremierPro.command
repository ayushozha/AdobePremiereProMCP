#!/bin/bash
# Double-click this file to launch the PremierPro AI Editor in Terminal.

cd "$(dirname "$0")"

# Check for API key
if [ -z "$ANTHROPIC_API_KEY" ]; then
    # Try loading from shell profile
    [ -f ~/.zshrc ] && source ~/.zshrc 2>/dev/null
    [ -f ~/.bashrc ] && source ~/.bashrc 2>/dev/null
    [ -f ~/.zprofile ] && source ~/.zprofile 2>/dev/null
    [ -f ~/.bash_profile ] && source ~/.bash_profile 2>/dev/null
fi

if [ -z "$ANTHROPIC_API_KEY" ]; then
    echo ""
    echo "  ❌ ANTHROPIC_API_KEY not found."
    echo ""
    echo "  Set it in your shell profile:"
    echo "    export ANTHROPIC_API_KEY=\"sk-ant-...\""
    echo ""
    echo "  Or paste it now:"
    read -rp "  API Key: " ANTHROPIC_API_KEY
    export ANTHROPIC_API_KEY
    echo ""
fi

# Ensure CLI dependencies are installed
if [ ! -d "cli/node_modules" ]; then
    echo "  Installing dependencies..."
    cd cli && npm install --silent && cd ..
fi

# Ensure MCP server binary exists
if [ ! -f "go-orchestrator/bin/premierpro-mcp" ]; then
    echo "  Building MCP server..."
    cd go-orchestrator && go build -o bin/premierpro-mcp ./cmd/server/ && cd ..
fi

# Launch the CLI
exec npx --prefix cli tsx cli/src/index.ts
