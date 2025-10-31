#!/bin/bash

echo "🔍 Checking GoVim Accessibility Permissions..."
echo ""

# Check if binary exists
if [ ! -f "bin/govim" ]; then
    echo "❌ Binary not found at bin/govim"
    echo "   Run: just build"
    exit 1
fi

echo "✓ Binary found: bin/govim"
echo ""

# Try to run a simple accessibility check
echo "Testing accessibility API access..."
./bin/govim status 2>&1 | head -5

echo ""
echo "📋 Manual Steps to Grant Permissions:"
echo ""
echo "1. Open System Settings"
echo "2. Go to Privacy & Security → Accessibility"
echo "3. Click the + button (unlock if needed)"
echo "4. Navigate to: $(pwd)/bin/"
echo "5. Select 'govim' and enable it"
echo ""
echo "OR use the app bundle:"
echo "   open GoVim.app"
echo ""
echo "Then restart GoVim and try the hotkeys again."
