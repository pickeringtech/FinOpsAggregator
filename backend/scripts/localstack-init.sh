#!/bin/bash
# LocalStack initialization script
# This runs automatically when LocalStack starts

set -e

echo "Initializing LocalStack S3 buckets..."

# Create import bucket
awslocal s3 mb s3://finops-imports || true
echo "Created bucket: finops-imports"

# Create export bucket
awslocal s3 mb s3://finops-exports || true
echo "Created bucket: finops-exports"

# Create bucket prefixes (by uploading empty marker files)
echo "" | awslocal s3 cp - s3://finops-imports/aws-cur/.keep
echo "" | awslocal s3 cp - s3://finops-imports/dynatrace/.keep
echo "" | awslocal s3 cp - s3://finops-exports/charts/.keep

echo "LocalStack S3 initialization complete!"

# List buckets to verify
echo "Available buckets:"
awslocal s3 ls

