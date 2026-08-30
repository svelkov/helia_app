#!/bin/bash

# PostgreSQL Migration Runner
# Usage: ./run_migrations.sh [username] [database] [host] [port]

USERNAME=${1:-postgres}
DATABASE=${2:-helia}
HOST=${3:-localhost}
PORT=${4:-5432}

echo "Running migrations for database: $DATABASE"
echo "User: $USERNAME"
echo "Host: $HOST:$PORT"
echo ""

# Check if psql is installed
if ! command -v psql &> /dev/null; then
    echo "Error: psql (PostgreSQL client) is not installed."
    exit 1
fi

# Run all migration files in order
for migration_file in migrations/00*.sql; do
    if [ -f "$migration_file" ]; then
        echo "Running migration: $migration_file"
        psql -U $USERNAME -d $DATABASE -h $HOST -p $PORT -f "$migration_file"
        
        if [ $? -eq 0 ]; then
            echo "✓ $migration_file completed successfully"
        else
            echo "✗ $migration_file failed"
            exit 1
        fi
        echo ""
    fi
done

echo "All migrations completed successfully!"
