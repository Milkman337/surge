#!/bin/bash
# install.sh - Install surge binary
# Usage: curl -sSL https://raw.githubusercontent.com/AtomicWasTaken/surge/main/scripts/install.sh | sh
set -e

echo "Installing surge via go install..."
go install github.com/AtomicWasTaken/surge/cmd/surge@latest

# Verify installation
if command -v surge &> /dev/null; then
    echo "Successfully installed surge $(surge --version)"
else
    echo "Installation failed - surge not found in PATH"
    echo "Make sure ~/go/bin is in your PATH"
    exit 1
fi
