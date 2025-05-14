#!/bin/bash
set -euo pipefail

echo "[$(date)] Starting database dump process..."

DUMP_FILE="/tmp/${DB_BACKUP_PREFIX}_backup_$(date +%Y%m%d_%H%M%S).sql.gz"

echo "[$(date)] Dumping database by connection string: ${DATABASE_URL}"
if pg_dump "${DATABASE_URL}" | gzip > "${DUMP_FILE}"; then
    echo "[$(date)] Database dump created successfully: ${DUMP_FILE}"
else
    echo "[$(date)] ERROR: Failed to create database dump"
    exit 1
fi

echo "[$(date)] Loading dump into DO Spaces: ${S3_BUCKET} via ${S3_ENDPOINT_URL}..."
if aws s3 cp "${DUMP_FILE}" "s3://${S3_BUCKET}/" --endpoint-url "${S3_ENDPOINT_URL}"; then
    echo "[$(date)] File successfully uploaded to DO Spaces"
else
    echo "[$(date)] ERROR: Uploading file to DO Spaces failed"
    exit 1
fi

rm -f "${DUMP_FILE}"
echo "[$(date)] Local dump file deleted"

echo "[$(date)] Backup completed successfully"
exit 0