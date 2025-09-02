#!/usr/bin/env bash

# Bokio Journal Entries Example Runner
# This script demonstrates how to use the pure generated API clients

echo "🚀 Bokio MCP - Journal Entries Example"
echo "======================================"

# Check if required environment variables are set
if [[ -z "$BOKIO_INTEGRATION_TOKEN" ]]; then
    echo "❌ Error: BOKIO_INTEGRATION_TOKEN environment variable is required"
    echo ""
    echo "📋 How to get your token:"
    echo "  1. Visit https://app.bokio.se"
    echo "  2. Go to Settings > Integrations"
    echo "  3. Create a new Integration Token"
    echo "  4. Export it: export BOKIO_INTEGRATION_TOKEN=\"your_token_here\""
    echo ""
    exit 1
fi

if [[ -z "$BOKIO_COMPANY_ID" ]]; then
    echo "❌ Error: BOKIO_COMPANY_ID environment variable is required"
    echo ""
    echo "📋 How to find your Company ID:"
    echo "  1. Check the URL when logged into Bokio web app"
    echo "  2. It's usually in the format: https://app.bokio.se/company/{COMPANY_ID}"
    echo "  3. Export it: export BOKIO_COMPANY_ID=\"your_company_uuid_here\""
    echo ""
    exit 1
fi

echo "✅ Environment variables configured:"
echo "   BOKIO_INTEGRATION_TOKEN: ${BOKIO_INTEGRATION_TOKEN:0:10}..."
echo "   BOKIO_COMPANY_ID: $BOKIO_COMPANY_ID"
echo "   BOKIO_BASE_URL: ${BOKIO_BASE_URL:-https://api.bokio.se (default)}"
echo ""

echo "🔨 Building example..."
if ! nix develop -c go build -o bin/list-journal-entries examples/list_journal_entries.go; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful!"
echo ""

echo "🚀 Running example..."
echo ""
./bin/list-journal-entries
