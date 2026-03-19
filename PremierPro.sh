#!/bin/bash
# Run this to launch the PremierPro AI Editor on Linux.
# Usage: ./PremierPro.sh

cd "$(dirname "$0")"

# Load shell profile for env vars
[ -f ~/.bashrc ] && source ~/.bashrc 2>/dev/null
[ -f ~/.bash_profile ] && source ~/.bash_profile 2>/dev/null
[ -f ~/.profile ] && source ~/.profile 2>/dev/null

# Try to resolve auth if no key is set
if [ -z "$ANTHROPIC_API_KEY" ] && [ -z "$OPENAI_API_KEY" ]; then
    if command -v claude &>/dev/null; then
        ANTHROPIC_API_KEY=$(claude auth print-api-key 2>/dev/null || true)
        [ -n "$ANTHROPIC_API_KEY" ] && export ANTHROPIC_API_KEY
    fi
    if [ -z "$ANTHROPIC_API_KEY" ] && command -v codex &>/dev/null; then
        OPENAI_API_KEY=$(codex auth print-api-key 2>/dev/null || true)
        [ -n "$OPENAI_API_KEY" ] && export OPENAI_API_KEY
    fi
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

if [ ! -d "cli/node_modules" ]; then
    echo "  Installing CLI dependencies..."
    cd cli && npm install --silent && cd ..
fi

if [ ! -d "ts-bridge/node_modules" ]; then
    echo "  Installing bridge dependencies..."
    cd ts-bridge && npm install --silent && cd ..
fi

if [ ! -f "go-orchestrator/bin/premierpro-mcp" ]; then
    echo "  Building MCP server..."
    cd go-orchestrator && go build -o bin/premierpro-mcp ./cmd/server/ && cd ..
fi

# Start backend services if not running
if ! lsof -i :50054 &>/dev/null 2>&1; then
    echo "  Starting backend services..."
    ./scripts/start-all.sh
    sleep 5
fi

exec npx --prefix cli tsx cli/src/index.ts
