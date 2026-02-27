set dotenv-load := true

project_name := 'savetoink'
lambda_archive := 'lambda-source.zip'
bucket_name := 'savetoink-lambda-source'
build_flags := "-X github.com/shaftoe/savetoink/internal/consts.version"

# Build CLI binary
build-cli:
    go build -ldflags "{{ build_flags }}=$(just version)" -o bin/savetoink ./cmd/cli

# Build CLI and convert URL into EPUB
run *ARGS: build-cli
    ./bin/savetoink convert {{ ARGS }}

# Run linter
lint:
    golangci-lint run

test: test-go test-webapp

# Run tests (skip DynamoDB integration tests)
test-go:
    go test ./... -short

# Build Lambda binary for Linux
build-lambda:
    GOOS=linux GOARCH=amd64 go build -ldflags "{{ build_flags }}=$(just version)" -o bin/bootstrap ./cmd/lambda

# Build Lambda zip for deployment
[working-directory('bin')]
build-lambda-zip: build-lambda
    zip {{ lambda_archive }} bootstrap

# Deploy S3 bucket for Lambda source
deploy-bucket:
    aws cloudformation deploy \
        --template-file infra/bucket.yaml \
        --stack-name {{ project_name }}-bucket \
        --parameter-overrides BucketName={{ bucket_name }}

# Deploy ACM certificate (must be deployed to us-east-1)
deploy-cert:
    @echo "Open https://us-east-1.console.aws.amazon.com/acm/certificates/ and add DNS validation records for $SAVETOINK_DOMAIN"
    aws cloudformation deploy \
        --template-file infra/cert.yaml \
        --stack-name {{ project_name }}-cert \
        --region us-east-1 \
        --parameter-overrides ProjectName={{ project_name }} DomainName="$SAVETOINK_DOMAIN"

# Get certificate ARN
get-cert-arn:
    aws cloudformation describe-stacks \
        --stack-name {{ project_name }}-cert \
        --region us-east-1 \
        --query "Stacks[0].Outputs[?OutputKey=='CertificateArn'].OutputValue" \
        --output text

# Upload Lambda source zip to S3
upload-zip:
    aws s3 cp bin/{{ lambda_archive }} s3://{{ bucket_name }}/{{ lambda_archive }}

# Deploy Lambda infrastructure
deploy-api:
    aws cloudformation deploy \
        --template-file infra/api.yaml \
        --stack-name {{ project_name }}-infra \
        --capabilities CAPABILITY_NAMED_IAM \
        --parameter-overrides \
            APIKeySecret="$SAVETOINK_API_KEY" \
            AppURL="$SAVETOINK_APP_URL" \
            Auth0Audience="$SAVETOINK_AUTH0_AUDIENCE" \
            Auth0Domain="$SAVETOINK_AUTH0_DOMAIN" \
            Auth0ClientId="$SAVETOINK_AUTH0_CLIENT_ID" \
            Auth0ClientSecret="$SAVETOINK_AUTH0_CLIENT_SECRET" \
            AuthBackend="auth0" \
            CertificateArn=$(just get-cert-arn) \
            DestEmail="$SAVETOINK_DEST_EMAIL" \
            DomainName="$SAVETOINK_DOMAIN" \
            MailjetAPIKey="$SAVETOINK_MAILJET_API_KEY" \
            MailjetAPISecret="$SAVETOINK_MAILJET_API_SECRET" \
            ProjectName={{ project_name }} \
            SenderEmail="$SAVETOINK_SENDER_EMAIL" \
            SourceBucketKey={{ lambda_archive }} \
            SourceBucketName={{ bucket_name }} \
            Debug="true"

# Full deployment (bucket + upload + infra)
deploy:
    just auth0-create-api
    just deploy-bucket
    just deploy-cert
    just build-lambda-zip
    just upload-zip
    just deploy-api
    @echo "Add DNS record: $SAVETOINK_DOMAIN A $(just get-distribution-url)."

# Destroy Lambda infrastructure
destroy:
    aws cloudformation delete-stack --stack-name {{ project_name }}-infra
    aws cloudformation wait stack-delete-complete --stack-name {{ project_name }}-infra
    -aws s3 rm s3://{{ bucket_name }} --recursive
    aws cloudformation delete-stack --stack-name {{ project_name }}-bucket
    aws cloudformation wait stack-delete-complete --stack-name {{ project_name }}-bucket
    aws cloudformation delete-stack --stack-name {{ project_name }}-cert --region us-east-1
    aws cloudformation wait stack-delete-complete --stack-name {{ project_name }}-cert --region us-east-1
    -just auth0-destroy-api

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

# View Lambda function logs
logs:
    aws logs tail /aws/lambda/{{ project_name }}-api --follow

# Test deployed Lambda function with article URL
test-create-article *URL:
    curl --silent \
      -X POST http://localhost:8080/v1/articles \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $SAVETOINK_API_KEY" \
      -d "{\"url\": \"{{ URL }}\", \"tags\":\"test\"}" | jq .

test-get-article *ID:
    curl --silent \
      -X GET http://localhost:8080/v1/articles/{{ ID }} \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $SAVETOINK_API_KEY" | jq .

test-get-articles *PAGE="1":
    curl --silent \
      -X GET http://localhost:8080/v1/articles?page={{ PAGE }} \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $SAVETOINK_API_KEY" | jq .

test-delete-article *ID:
    curl --silent \
      -X DELETE http://localhost:8080/v1/articles/{{ ID }} \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $SAVETOINK_API_KEY" | jq .

test-auth-token-exchange *CODE:
    curl --silent \
        -X POST \
        --data '{"code":"{{ CODE }}", "redirect_uri": "http://localhost:5173"}' \
        http://localhost:8080/v1/auth/token | jq .

test-auth-browser-login:
    open "https://${SAVETOINK_AUTH0_DOMAIN}/authorize?response_type=code&client_id=${SAVETOINK_AUTH0_CLIENT_ID}&redirect_uri=http://localhost:5173&scope=openid%20profile%20email&state=test123&audience=${SAVETOINK_AUTH0_AUDIENCE}"

deploy-lambda: build-lambda-zip upload-zip
    aws lambda update-function-code \
        --function-name {{ project_name }}-api \
        --s3-bucket {{ bucket_name }} \
        --s3-key {{ lambda_archive }} \
        --publish

server-http:
    reflex -r '\.(env|go)$' -s -- go run -ldflags "{{ build_flags }}=$(just version)" ./cmd/http/main.go

[working-directory('cmd/webapp')]
server-webapp:
    bun run dev --open

[working-directory('cmd/website')]
server-website:
    npm run dev

[working-directory('cmd/webapp')]
test-webapp:
    npm run check
    npm run lint
    npm run test

# Scan DynamoDB article table and print all records
scan-table TABLE_NAME="savetoink-articles":
    aws dynamodb scan \
        --table-name {{ TABLE_NAME }} \
        --output json \
        --query 'Items[*].{ID:id.S,URL:url.S,Title:title.S,Author:author.S,Status:deliveryStatus.S,Created:createdAt.S}'

auth0-create-api:
    auth0 apis create \
    --name {{ project_name }} \
    --identifier "$SAVETOINK_AUTH0_AUDIENCE" \
    --signing-alg "RS256"

auth0-destroy-api:
    auth0 apis delete --force "$SAVETOINK_AUTH0_AUDIENCE"

version:
    @echo "$(cat VERSION)-$(date -u +%Y%m%d)-$(git rev-parse --short HEAD)"

upgrade-go-deps:
    go get -u all

[working-directory('cmd/webapp')]
upgrade-svelte-deps:
    npm upgrade

upgrade-deps: upgrade-go-deps upgrade-svelte-deps

dump-table TABLE_NAME:
    uvx dynamodump -r $AWS_REGION -m backup -s {{ TABLE_NAME }}

restore-table TABLE_NAME:
    uvx dynamodump -r $AWS_REGION -m restore -s {{ TABLE_NAME }}

bootstrap-website:
    npm create astro@latest cmd/website

boostrap-extension:
    bunx wxt@latest init
