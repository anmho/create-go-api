#!/bin/bash
set -e

echo "Generating code..."

{{- if .HasConnectRPC}}
echo "Generating protobuf code..."
echo "Updating buf dependencies..."
buf dep update
echo "Generating code from protobuf definitions..."
buf generate
{{- end}}

echo "Generating mocks..."
go generate ./...

echo "✓ Code generation complete"

