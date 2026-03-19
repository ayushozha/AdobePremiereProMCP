#!/bin/bash
# Double-click this file to launch the PremierPro AI Editor in Terminal.

cd "$(dirname "$0")"

# Load shell profile for env vars
[ -f ~/.zshrc ] && source ~/.zshrc 2>/dev/null
[ -f ~/.zprofile ] && source ~/.zprofile 2>/dev/null
[ -f ~/.bashrc ] && source ~/.bashrc 2>/dev/null
[ -f ~/.bash_profile ] && source ~/.bash_profile 2>/dev/null

# Try to resolve auth if no key is set
if [ -z "$ANTHROPIC_API_KEY" ] && [ -z "$OPENAI_API_KEY" ]; then
    # Try Claude Code auth
    if command -v claude &>/dev/null; then
        ANTHROPIC_API_KEY=$(claude auth print-api-key 2>/dev/null || true)
        [ -n "$ANTHROPIC_API_KEY" ] && export ANTHROPIC_API_KEY
    fi
    # Try Codex auth
    if [ -z "$ANTHROPIC_API_KEY" ] && command -v codex &>/dev/null; then
        OPENAI_API_KEY=$(codex auth print-api-key 2>/dev/null || true)
        [ -n "$OPENAI_API_KEY" ] && export OPENAI_API_KEY
    fi
    # If still nothing, offer to login
    if [ -z "$ANTHROPIC_API_KEY" ] && [ -z "$OPENAI_API_KEY" ]; then
        echo ""
        echo "  No API key found. Choose a provider to login:"
        echo "    1) Claude (Anthropic)  — claude login"
        echo "    2) OpenAI / Codex      — codex login"
        echo ""
        read -rp "  Choice [1]: " choice
        case "$choice" in
            2)
                if command -v codex &>/dev/null; then
                    codex login
                    OPENAI_API_KEY=$(codex auth print-api-key 2>/dev/null || true)
                    export OPENAI_API_KEY
                else
                    echo "  codex CLI not found. Install: npm install -g @openai/codex"
                    read -rp "  Or paste your OpenAI API key: " OPENAI_API_KEY
                    export OPENAI_API_KEY
                fi
                ;;
            *)
                if command -v claude &>/dev/null; then
                    claude login
                    ANTHROPIC_API_KEY=$(claude auth print-api-key 2>/dev/null || true)
                    export ANTHROPIC_API_KEY
                else
                    echo "  claude CLI not found. Install: npm install -g @anthropic-ai/claude-code"
                    read -rp "  Or paste your Anthropic API key: " ANTHROPIC_API_KEY
                    export ANTHROPIC_API_KEY
                fi
                ;;
        esac
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

exec npx --prefix cli tsx cli/src/index.ts
