#!/bin/bash
# Run Lambda functions locally using SAM CLI
set -e

cd "$(dirname "$0")/.."

FUNCTION=${1:-"ImportAWSCURFunction"}
EVENT_FILE=${2:-"events/s3-put.json"}

echo "Invoking Lambda function locally: $FUNCTION"

# Ensure the function is built
if [ ! -f ".aws-sam/build/$FUNCTION/bootstrap" ]; then
    echo "Building Lambda functions first..."
    ./scripts/sam-build.sh
fi

# Create events directory if it doesn't exist
mkdir -p events

# Create sample S3 event if it doesn't exist
if [ ! -f "events/s3-put.json" ]; then
    cat > events/s3-put.json << 'EOF'
{
  "Records": [
    {
      "eventVersion": "2.1",
      "eventSource": "aws:s3",
      "awsRegion": "eu-west-1",
      "eventTime": "2024-01-15T12:00:00.000Z",
      "eventName": "ObjectCreated:Put",
      "s3": {
        "s3SchemaVersion": "1.0",
        "bucket": {
          "name": "finops-import-dev",
          "arn": "arn:aws:s3:::finops-import-dev"
        },
        "object": {
          "key": "aws-cur/sample_aws_cur.csv",
          "size": 1024
        }
      }
    }
  ]
}
EOF
fi

# Create sample allocate event if it doesn't exist
if [ ! -f "events/allocate.json" ]; then
    cat > events/allocate.json << 'EOF'
{
  "start_date": "2024-01-01",
  "end_date": "2024-01-31"
}
EOF
fi

# Create sample export event if it doesn't exist
if [ ! -f "events/export.json" ]; then
    cat > events/export.json << 'EOF'
{
  "type": "chart_graph",
  "format": "png",
  "date": "2024-01-15"
}
EOF
fi

# Invoke the function
sam local invoke "$FUNCTION" \
    --event "$EVENT_FILE" \
    --env-vars env.json \
    --docker-network host

echo "Invocation complete!"

