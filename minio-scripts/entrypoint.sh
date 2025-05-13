#!/bin/sh
set -e

# Original MinIO command: server /data --console-address ":9001"
MINIO_CMD="minio server /data --console-address :9001"

echo "Starting MinIO server in background: $MINIO_CMD"
$MINIO_CMD &
MINIO_PID=$!

echo "Waiting for MinIO server (http://localhost:9000) to become healthy..."
MAX_HEALTH_ATTEMPTS=30
HEALTH_ATTEMPT_COUNT=0
# Use curl to check the health endpoint.
# The -s flag is for silent, -f for fail fast (non-zero exit on server errors).
until curl -sf "http://localhost:9000/minio/health/ready"; do
    HEALTH_ATTEMPT_COUNT=$((HEALTH_ATTEMPT_COUNT + 1))
    if [ "$HEALTH_ATTEMPT_COUNT" -ge "$MAX_HEALTH_ATTEMPTS" ]; then
        echo "MinIO server did not become healthy after $MAX_HEALTH_ATTEMPTS attempts. Terminating."
        kill $MINIO_PID # Attempt to kill the background MinIO server
        wait $MINIO_PID 2>/dev/null # Wait for it to terminate
        exit 1
    fi
    echo "MinIO not ready yet (health check attempt $HEALTH_ATTEMPT_COUNT/$MAX_HEALTH_ATTEMPTS). Retrying in 3 seconds..."
    sleep 3
done
echo "MinIO server is healthy and running."

echo "Executing MinIO setup script (/minio-scripts/setup_minio.sh)..."
# Ensure the setup script is executable
chmod +x /minio-scripts/setup_minio.sh
# Execute the script that configures mc and sets bucket policies
sh /minio-scripts/setup_minio.sh

echo "MinIO setup script finished."
echo "MinIO service is fully configured and running. PID: $MINIO_PID"

# Wait for the MinIO server process to exit.
# This keeps the container running as long as MinIO server is running.
wait $MINIO_PID