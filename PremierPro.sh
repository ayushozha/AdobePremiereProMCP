#!/bin/bash
# Run this to launch the PremierPro AI Editor on Linux.
# Usage: ./PremierPro.sh

cd "$(dirname "$0")"

# Check for API key from env or shell profile
if [ -z "$ANTHROPIC_API_KEY" ]; then
    [ -f ~/.bashrc ] && source ~/.bashrc 2>/dev/null
    [ -f ~/.bash_profile ] && source ~/.bash_profile 2>/dev/null
    [ -f ~/.profile ] && source ~/.profile 2>/dev/null
fi

# If still no key, try Claude Code auth
if [ -z "$ANTHROPIC_API_KEY" ] && command -v claude &>/dev/null; then
    ANTHROPIC_API_KEY=$(claude auth print-api-key 2>/dev/null || true)
    export ANTHROPIC_API_KEY
fi

# If still no key, offer to login via Claude Code
if [ -z "$ANTHROPIC_API_KEY" ]; then
    if command -v claude &>/dev/null; then
        echo ""
        echo "  No API key found. Launching Claude login..."
        echo ""
        claude login
        ANTHROPIC_API_KEY=$(claude auth print-api-key 2>/dev/null || true)
        export ANTHROPIC_API_KEY
    fi
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
