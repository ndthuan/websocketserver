#!/usr/bin/env bash

set -e

if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
  echo "Bad code format"
  exit 1
fi

echo "=== unit test ==="

go test

echo "=== functional test ==="

go build -o /tmp/chatserver ./examples/chat-server/

/tmp/chatserver > /tmp/output 2>&1 &
SERVER_PID=$!
go run examples/chat-client/main.go >> /tmp/output 2>&1 &
go run examples/chat-client/main.go >> /tmp/output 2>&1 &

sleep 10

kill $SERVER_PID

grep -q "http server started on" /tmp/output || (echo "server not started"; exit 1)
grep -q "Hi there, this is private message to you" /tmp/output || (echo "private message not sent"; exit 1)
grep -q "Everyone, please welcome" /tmp/output || (echo "broadcast messages not sent"; exit 1)
grep -q "how are you" /tmp/output || (echo "standalone runner's messages not sent"; exit 1)

echo "All good"
