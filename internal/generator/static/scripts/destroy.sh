#!/bin/bash
set -e

PROJECT_NAME="{{.ProjectName}}"

echo "⚠️  WARNING: This will permanently delete the Fly.io app '$PROJECT_NAME' and all associated resources!"
echo "This action cannot be undone."
read -p "Are you sure you want to continue? (type 'yes' to confirm): " confirm

if [ "$confirm" != "yes" ]; then
    echo "Destroy cancelled."
    exit 1
fi

echo "Destroying Fly.io app..."

if ! command -v flyctl >/dev/null 2>&1 && ! command -v fly >/dev/null 2>&1; then
    echo "Error: flyctl or fly command not found. Install from https://fly.io/docs/getting-started/installing-flyctl/"
    exit 1
fi

if command -v flyctl >/dev/null 2>&1; then
    flyctl apps destroy "$PROJECT_NAME" --yes
elif command -v fly >/dev/null 2>&1; then
    fly apps destroy "$PROJECT_NAME" --yes
fi

echo "✓ App destroyed"

