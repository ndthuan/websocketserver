#!/usr/bin/env bash

set -e

if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
  echo "Bad code format"
  exit 1
fi

go test
