# Scan DynamoDB article table and print all records
scan-table TABLE_NAME="savetoink-articles":
    aws dynamodb scan \
        --table-name {{ TABLE_NAME }} \
        --output json \
        --query 'Items[*].{ID:id.S,URL:url.S,Title:title.S,Author:author.S,Status:deliveryStatus.S,Created:createdAt.S}'

dump-table TABLE_NAME:
    uvx dynamodump -r $AWS_REGION -m backup -s {{ TABLE_NAME }}

restore-table TABLE_NAME:
    uvx dynamodump -r $AWS_REGION -m restore -s {{ TABLE_NAME }}
