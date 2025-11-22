#!/usr/bin/env bash

install() {
  echo "Installing Git hooks..."

  echo "=> Added .git/hooks/pre-push"
  cp scripts/git/hooks/pre-push .git/hooks/pre-push
  chmod +x .git/hooks/pre-push

  echo "Hooks installed."
}