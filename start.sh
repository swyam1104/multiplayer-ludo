#!/bin/bash

echo "Starting Multiplayer Ludo Backend..."

# Environment variables for the local native setup
export DB_DSN="ludo_user:ludopassword@tcp(127.0.0.1:3307)/ludo?parseTime=true"
export JWT_SECRET="supersecretjwtkey"
export PORT="8080"

# Run the Go server
go run ./cmd/server
