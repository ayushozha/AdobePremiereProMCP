#!/bin/bash
# Run this to launch the PremierPro AI Editor on Linux.
# Usage: ./PremierPro.sh

cd "$(dirname "$0")"

# Check for API key
if [ -z "$ANTHROPIC_API_KEY" ]; then
    [ -f ~/.bashrc ] && source ~/.bashrc 2>/dev/null
    [ -f ~/.bash_profile ] && source ~/.bash_profile 2>/dev/null
    [ -f ~/.profile ] && source ~/.profile 2>/dev/null
fi

if [ -z "$ANTHROPIC_API_KEY" ]; then
    echo ""
    echo "  ANTHROPIC_API_KEY not found."
    echo ""
    echo "  Set it in your shell profile:"
    echo "    export ANTHROPIC_API_KEY=\"sk-ant-...\""
    echo ""
    printf "  API Key: "
    read -r ANTHROPIC_API_KEY
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
