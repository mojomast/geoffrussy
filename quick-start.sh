#!/bin/bash

# Quick Start Test Script - Run Geoffrey's core workflow
# This script initializes Geoffrey and runs a quick test of the interview

set -e

echo "╔════════════════════════════════════════════════════════════╗"
echo "║              Geoffrey Quick Start Test                           ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Build Geoffrey
echo "🔨 Building Geoffrey..."
go build ./cmd/geoffrussy
echo "✅ Build complete!"
echo ""

# Initialize
echo "📝 Step 1: Initialize Geoffrey"
echo "   You will be prompted for API keys"
echo ""
./geoffrussy init
echo ""

echo "✅ Geoffrey is ready!"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Next steps:"
echo "  1. Run './geoffrussy interview' to start gathering requirements"
echo "  2. Run './geoffrussy design' to generate architecture"
echo "  3. Run './geoffrussy plan' to create development plan"
echo "  4. Run './geoffrussy review' to review the plan"
echo "  5. Run './geoffrussy develop' to execute development"
echo ""
echo "Or run './test-interactive.sh' for an interactive menu!"
echo ""
