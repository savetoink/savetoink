# Get Lambda function URL
get-url:
    aws cloudformation describe-stacks \
        --stack-name {{ project_name }}-infra \
        --query "Stacks[0].Outputs[?OutputKey=='FunctionUrl'].OutputValue" \
        --output text

# Get CloudFront distribution domain name
get-distribution-url:
    aws cloudformation describe-stacks \
        --stack-name {{ project_name }}-infra \
        --query "Stacks[0].Outputs[?OutputKey=='CloudFrontDomainName'].OutputValue" \
        --output text

# Get GitHub Actions IAM role ARN
get-github-role:
    aws cloudformation describe-stacks \
        --stack-name {{ project_name }}-infra \
        --query "Stacks[0].Outputs[?OutputKey=='GitHubActionsRoleArn'].OutputValue" \
        --output text

# View API Lambda function logs
logs-api:
    aws logs tail /aws/lambda/{{ project_name }}-api --follow

logs-processor:
    aws logs tail /aws/lambda/{{ project_name }}-processor --follow

# Get certificate ARN
get-cert-arn:
    aws cloudformation describe-stacks \
        --stack-name {{ project_name }}-cert \
        --region us-east-1 \
        --query "Stacks[0].Outputs[?OutputKey=='CertificateArn'].OutputValue" \
        --output text
