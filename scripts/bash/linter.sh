#!/bin/bash

output="$(./custom-gcl run -c .golangci.yml 2>&1)"
status="$?"
echo "$output"

if [ "$status" -ne 0 ]; then
  if echo "$output" | grep -Eq "build linters|module.* not found"; then
    echo "❌ GolangCI-Lint plugin build failed. Try 'make linter-rebuild'."
  fi
  exit "$status"
fi

gci list . | sed 's/^/BadFormat: /'
[ "$(gci list . | wc -c)" -eq 0 ]