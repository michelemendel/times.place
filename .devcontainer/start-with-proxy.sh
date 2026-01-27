#!/bin/bash
# Start socat proxy to forward port 5432 to postgres:5432
# This allows Cursor's port forwarding to work for pgAdmin and other GUI tools
socat TCP-LISTEN:5432,fork,reuseaddr TCP:postgres:5432 &
SOCAT_PID=$!

# Keep the container running
sleep infinity

# Cleanup on exit
kill $SOCAT_PID 2>/dev/null
