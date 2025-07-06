#!/bin/bash
# Quick install script for Peaks

set -e

echo "🏔️  Installing Peaks - Beautiful Terminal Bandwidth Monitor"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or higher first."
    echo "   Visit: https://golang.org/dl/"
    echo ""
    echo "Alternative: Download a pre-built binary from:"
    echo "   https://github.com/marcodenic/peaks/releases"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.21"

if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "❌ Go version $GO_VERSION is too old. Please upgrade to Go 1.21 or higher."
    echo ""
    echo "Alternative: Download a pre-built binary from:"
    echo "   https://github.com/marcodenic/peaks/releases"
    exit 1
fi

echo "✅ Go version $GO_VERSION found"

# Install Peaks
echo "📦 Installing Peaks..."
if go install github.com/marcodenic/peaks/cmd/peaks@latest; then
    echo "✅ Peaks installed successfully!"
    echo ""
    echo "🚀 You can now run: peaks"
    echo ""
    echo "📖 For help and usage information:"
    echo "   https://github.com/marcodenic/peaks"
else
    echo "❌ Installation failed!"
    echo ""
    echo "Alternative: Download a pre-built binary from:"
    echo "   https://github.com/marcodenic/peaks/releases"
    exit 1
fi
    echo ""
    echo "🚀 Run 'peaks' to start monitoring your bandwidth!"
    echo ""
    echo "💡 Controls:"
    echo "   q/Ctrl+C - Quit"
    echo "   p/Space  - Pause/Resume"
    echo "   r        - Reset chart"
    echo "   s        - Toggle statusbar"
else
    echo "❌ Installation failed. Please check your internet connection and try again."
    exit 1
fi
