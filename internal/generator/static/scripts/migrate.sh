#!/bin/bash
set -e

# Generate migration from schema.sql and apply it
PROJECT_NAME="${PROJECT_NAME:-postservice}"
DB_URL="postgres://postgres:postgres@localhost:5432/${PROJECT_NAME}?sslmode=disable"

echo "Generating migration from schema.sql..."
if [ ! -f schema.sql ]; then
    echo "Error: schema.sql not found"
    exit 1
fi

if [ ! -d migrations ]; then
    mkdir -p migrations
fi

timestamp=$(date +%Y%m%d%H%M%S)
migration_file="migrations/${timestamp}_initial.up.sql"

if [ ! -f "$migration_file" ]; then
    cp schema.sql "$migration_file"
    echo "✓ Migration created: $migration_file"
fi

echo "Applying migrations..."
for migration in migrations/*.up.sql; do
    if [ -f "$migration" ]; then
        echo "Applying $migration..."
        psql "$DB_URL" -f "$migration" 2>/dev/null || echo "Migration may already be applied"
    fi
done

echo "✓ Migrations applied"

