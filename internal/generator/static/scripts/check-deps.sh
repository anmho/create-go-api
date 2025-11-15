#!/bin/bash
set -e

echo "Installing dependencies..."

{{- if .HasConnectRPC}}
echo "Checking buf..."
if ! command -v buf >/dev/null 2>&1; then
    echo "buf is not installed. Install from https://buf.build/docs/installation"
    exit 1
fi
echo "✓ buf installed"
{{- end}}

{{- if .HasPostgres}}
echo "Checking psql..."
if command -v psql >/dev/null 2>&1; then
    echo "✓ psql already installed"
else
    echo "Error: psql is not installed. Install PostgreSQL client tools:"
    echo "  macOS: brew install postgresql"
    echo "  Linux: apt-get install postgresql-client"
    exit 1
fi
{{- end}}

echo "✓ All dependencies installed"

