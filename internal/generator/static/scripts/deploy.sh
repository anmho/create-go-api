#!/bin/bash
set -e

PROJECT_NAME="{{.ProjectName}}"

if ! command -v flyctl >/dev/null 2>&1 && ! command -v fly >/dev/null 2>&1; then
    echo "Error: flyctl or fly command not found. Install from https://fly.io/docs/getting-started/installing-flyctl/"
    exit 1
fi

if command -v flyctl >/dev/null 2>&1; then
    flyctl deploy -a "$PROJECT_NAME"
elif command -v fly >/dev/null 2>&1; then
    fly deploy -a "$PROJECT_NAME"
fi

