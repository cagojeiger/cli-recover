#!/bin/bash

# Test script to verify TUI backup type functionality

echo "Building cli-recover..."
go build -o cli-recover ./cmd/cli-recover

echo "Testing TUI backup type screen..."
echo "You should see:"
echo "1. Main menu -> select Backup"
echo "2. Backup type selection screen with 3 options:"
echo "   - filesystem"
echo "   - minio" 
echo "   - mongodb"
echo "3. After selecting a type, you should see namespace list"
echo ""
echo "Press Ctrl+C to exit at any time"
echo ""

./cli-recover tui

echo "Test completed!"