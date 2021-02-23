#!/usr/bin/env bash

set -ex

if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
  echo "Bad code format"
  exit 1
fi

echo "=== mod download ==="

go mod download

echo "=== unit test ==="

go test

echo "=== functional test ==="

go build -o /tmp/chatserver ./examples/chat-server/
go build -o /tmp/chatclient ./examples/chat-client/

/tmp/chatserver > /tmp/output 2>&1 &
SERVER_PID=$!

sleep 3

/tmp/chatclient > /tmp/client-output 2>&1 &
/tmp/chatclient >> /tmp/client-output 2>&1 &
/tmp/chatclient >> /tmp/client-output 2>&1 &

sleep 10

kill $SERVER_PID

grep -q "http server started on" /tmp/output || (echo "server was not started"; exit 1)

clientsConnected=$(grep -c 'Client connected' /tmp/output)
if [ "$clientsConnected" != "3" ]; then
  echo "there must be 3 clients connected, found: $clientsConnected"
  exit 1
fi

standaloneStarted=$(grep -c "Standalone runner was started" /tmp/output)
if [ "$standaloneStarted" != "1" ]; then
  echo "standalone runner should be started only once, found: $standaloneStarted"
  exit 1
fi

grep -q "Login handler triggered" /tmp/output || (echo "login handler was not triggered"; exit 1)
grep -q "Logout handler triggered" /tmp/output || (echo "logout handler was not triggered"; exit 1)
grep -q "Client disconnected" /tmp/output || (echo "disconnection callback was not triggered"; exit 1)

echo "All good"
