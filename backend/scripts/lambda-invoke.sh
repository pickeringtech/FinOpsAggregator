#!/bin/bash
# Invoke Lambda functions locally via Docker Compose
# Usage:
#   ./scripts/lambda-invoke.sh import-awscur [file.csv]
#   ./scripts/lambda-invoke.sh import-dynatrace [file.json]
#   ./scripts/lambda-invoke.sh allocate [start_date] [end_date]
#   ./scripts/lambda-invoke.sh export [type] [format]

set -e

cd "$(dirname "$0")/.."

COMMAND=${1:-"help"}
LOCALSTACK_ENDPOINT="http://localhost:4566"

# Check if LocalStack is running
check_localstack() {
    if ! curl -s "$LOCALSTACK_ENDPOINT/_localstack/health" > /dev/null 2>&1; then
        echo "Error: LocalStack is not running. Start it with:"
        echo "  docker compose -f docker-compose.lambda.yml up -d"
        exit 1
    fi
}

# Upload a file to LocalStack S3
upload_to_s3() {
    local file=$1
    local bucket=$2
    local key=$3
    
    echo "Uploading $file to s3://$bucket/$key..."
    AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test \
        aws --endpoint-url="$LOCALSTACK_ENDPOINT" \
        s3 cp "$file" "s3://$bucket/$key"
}

# Invoke Lambda via Docker exec
invoke_lambda() {
    local handler=$1
    local event=$2
    
    echo "Invoking Lambda handler: $handler"
    echo "Event: $event"
    
    docker compose -f docker-compose.lambda.yml exec -T lambda \
        sh -c "FINOPS_LAMBDA_HANDLER=$handler echo '$event' | /app/bootstrap"
}

# Run Lambda binary directly in container
run_lambda() {
    local handler=$1
    shift
    local env_vars="$@"
    
    echo "Running Lambda handler: $handler"
    
    docker compose -f docker-compose.lambda.yml run --rm \
        -e "FINOPS_LAMBDA_HANDLER=$handler" \
        $env_vars \
        lambda /app/bootstrap
}

case "$COMMAND" in
    import-awscur)
        check_localstack
        FILE=${2:-"testdata/sample_aws_cur.csv"}
        
        if [ ! -f "$FILE" ]; then
            echo "Error: File not found: $FILE"
            exit 1
        fi
        
        # Upload file to LocalStack S3
        FILENAME=$(basename "$FILE")
        upload_to_s3 "$FILE" "finops-imports" "aws-cur/$FILENAME"
        
        # Create S3 event
        EVENT=$(cat <<EOF
{
  "Records": [
    {
      "eventVersion": "2.1",
      "eventSource": "aws:s3",
      "awsRegion": "eu-west-1",
      "eventTime": "$(date -u +%Y-%m-%dT%H:%M:%S.000Z)",
      "eventName": "ObjectCreated:Put",
      "s3": {
        "s3SchemaVersion": "1.0",
        "bucket": {
          "name": "finops-imports",
          "arn": "arn:aws:s3:::finops-imports"
        },
        "object": {
          "key": "aws-cur/$FILENAME",
          "size": $(stat -f%z "$FILE" 2>/dev/null || stat -c%s "$FILE")
        }
      }
    }
  ]
}
EOF
)
        
        run_lambda "import_awscur" -e "FINOPS_IMPORT_SOURCE=aws_cur"
        ;;
        
    import-dynatrace)
        check_localstack
        FILE=${2:-""}
        NODE_ID=${3:-""}
        
        if [ -z "$FILE" ]; then
            echo "Usage: $0 import-dynatrace <file.json> [node-id]"
            exit 1
        fi
        
        if [ ! -f "$FILE" ]; then
            echo "Error: File not found: $FILE"
            exit 1
        fi
        
        # Use node ID from argument or extract from filename
        if [ -z "$NODE_ID" ]; then
            NODE_ID=$(basename "$FILE" .json)
        fi
        
        # Upload file to LocalStack S3
        upload_to_s3 "$FILE" "finops-imports" "dynatrace/$NODE_ID.json"
        
        run_lambda "import_dynatrace" -e "FINOPS_IMPORT_SOURCE=dynatrace"
        ;;
        
    allocate)
        check_localstack
        START_DATE=${2:-$(date -d "30 days ago" +%Y-%m-%d 2>/dev/null || date -v-30d +%Y-%m-%d)}
        END_DATE=${3:-$(date +%Y-%m-%d)}
        
        echo "Running allocation for period: $START_DATE to $END_DATE"
        
        # For allocate, we need to pass the request as stdin
        EVENT=$(cat <<EOF
{
  "start_date": "$START_DATE",
  "end_date": "$END_DATE"
}
EOF
)
        
        echo "$EVENT" | docker compose -f docker-compose.lambda.yml run --rm \
            -e "FINOPS_LAMBDA_HANDLER=allocate" \
            -T lambda /app/bootstrap
        ;;
        
    export)
        check_localstack
        TYPE=${2:-"chart_graph"}
        FORMAT=${3:-"png"}
        DATE=${4:-$(date +%Y-%m-%d)}
        
        echo "Exporting: type=$TYPE, format=$FORMAT, date=$DATE"
        
        EVENT=$(cat <<EOF
{
  "type": "$TYPE",
  "format": "$FORMAT",
  "date": "$DATE"
}
EOF
)
        
        echo "$EVENT" | docker compose -f docker-compose.lambda.yml run --rm \
            -e "FINOPS_LAMBDA_HANDLER=export" \
            -T lambda /app/bootstrap
        ;;
        
    list-s3)
        check_localstack
        BUCKET=${2:-"finops-imports"}
        
        echo "Listing contents of s3://$BUCKET..."
        AWS_ACCESS_KEY_ID=test AWS_SECRET_ACCESS_KEY=test \
            aws --endpoint-url="$LOCALSTACK_ENDPOINT" \
            s3 ls "s3://$BUCKET/" --recursive
        ;;
        
    shell)
        echo "Opening shell in Lambda container..."
        docker compose -f docker-compose.lambda.yml exec lambda /bin/bash
        ;;
        
    logs)
        docker compose -f docker-compose.lambda.yml logs -f lambda
        ;;
        
    help|*)
        echo "Lambda Local Development Helper"
        echo ""
        echo "Usage: $0 <command> [args...]"
        echo ""
        echo "Commands:"
        echo "  import-awscur [file.csv]           Import AWS CUR data"
        echo "  import-dynatrace <file.json> [id]  Import Dynatrace metrics"
        echo "  allocate [start_date] [end_date]   Run cost allocation"
        echo "  export [type] [format] [date]      Export charts/data"
        echo "  list-s3 [bucket]                   List S3 bucket contents"
        echo "  shell                              Open shell in Lambda container"
        echo "  logs                               Follow Lambda container logs"
        echo ""
        echo "Export types: chart_graph, chart_trend, chart_waterfall"
        echo "Export formats: png, svg"
        echo ""
        echo "Examples:"
        echo "  $0 import-awscur testdata/sample_aws_cur.csv"
        echo "  $0 allocate 2024-01-01 2024-01-31"
        echo "  $0 export chart_graph png 2024-01-15"
        ;;
esac

