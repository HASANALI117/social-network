#!/bin/sh
set -e # Exit immediately if a command exits with a non-zero status.

# Environment variables MINIO_ROOT_USER and MINIO_ROOT_PASSWORD are expected to be set by Docker Compose.
ACCESS_KEY="${MINIO_ROOT_USER}"
SECRET_KEY="${MINIO_ROOT_PASSWORD}"
BUCKET_NAME="images" # The bucket to make public
ALIAS_NAME="local"   # Alias for the local MinIO server
MINIO_API_ENDPOINT="http://localhost:9000" # MinIO server endpoint within the container network

echo "Attempting to configure MinIO client (mc) alias..."
# Loop to wait for mc to be ready and server to be responsive for alias creation
MAX_MC_ATTEMPTS=10
MC_ATTEMPT_COUNT=0
# The --api "S3v4" flag is often important for compatibility.
until mc alias set "${ALIAS_NAME}" "${MINIO_API_ENDPOINT}" "${ACCESS_KEY}" "${SECRET_KEY}" --api "S3v4"; do
    MC_ATTEMPT_COUNT=$((MC_ATTEMPT_COUNT + 1))
    if [ "$MC_ATTEMPT_COUNT" -ge "$MAX_MC_ATTEMPTS" ]; then
        echo "Failed to set mc alias after $MAX_MC_ATTEMPTS attempts. Please check MinIO server status and credentials."
        exit 1
    fi
    echo "mc alias set failed (Attempt $MC_ATTEMPT_COUNT/$MAX_MC_ATTEMPTS). Retrying in 3 seconds..."
    sleep 3
done
echo "MinIO client (mc) alias '${ALIAS_NAME}' configured successfully."

echo "Checking if bucket '${BUCKET_NAME}' exists..."
if ! mc ls "${ALIAS_NAME}/${BUCKET_NAME}" > /dev/null 2>&1; then
  echo "Bucket '${BUCKET_NAME}' does not exist. Creating it..."
  mc mb "${ALIAS_NAME}/${BUCKET_NAME}"
  echo "Bucket '${BUCKET_NAME}' created."
else
  echo "Bucket '${BUCKET_NAME}' already exists."
fi

echo "Setting access policy for bucket '${BUCKET_NAME}' to public read-only..."
# MINIO_DEFAULT_BUCKETS should have created the 'images' bucket.
# 'mc policy set download' makes the bucket contents publicly readable.
mc anonymous set download "${ALIAS_NAME}/${BUCKET_NAME}"

echo "Access policy for bucket '${BUCKET_NAME}' successfully set to public read-only."