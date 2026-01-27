#!/bin/bash
set -e

# Create test database (ignore error if it already exists)
psql -v ON_ERROR_STOP=0 --username "$POSTGRES_USER" --dbname "postgres" -c "CREATE DATABASE timesplace_test;" || true

echo "Test database 'timesplace_test' ready"
