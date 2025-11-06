#!/bin/bash

echo "🧪 Running Backend Auth Tests"

# Start MongoDB if not running
if ! pgrep -x "mongod" > /dev/null; then
    echo "📦 Starting MongoDB..."
    brew services start mongodb/brew/mongodb-community
    sleep 3
fi

# Run tests
echo "🔧 Running auth flow tests..."
cd /Users/kwabena/Documents/project_files/healthyPay/backend
go test ./test/... -v

echo "✅ Tests completed!"
