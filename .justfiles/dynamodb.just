# Scan DynamoDB article table and print all records
scan-table TABLE_NAME="savetoink-articles":
    aws dynamodb scan \
        --table-name {{ TABLE_NAME }} \
        --output json \
        --query 'Items[*].{ID:id.S,URL:url.S,Title:title.S,Author:author.S,Status:deliveryStatus.S,Created:createdAt.S}'

backup:
    @aws lambda invoke \
        --function-name {{ project_name }}-scheduler \
        --cli-binary-format raw-in-base64-out \
        --payload '{"task": "backup"}' \
        /dev/stdout \
        | jq .

restore backup_name:
    @aws lambda invoke \
        --function-name "{{ project_name }}-scheduler" \
        --cli-binary-format raw-in-base64-out \
        --payload '{"task": "restore", "backup_name": "{{ backup_name }}", "overwrite": true}' \
        /dev/stdout \
        | jq .

list-backups:
    @aws s3 ls "${SAVETOINK_BACKUP_BUCKET_NAME}/backups/"
