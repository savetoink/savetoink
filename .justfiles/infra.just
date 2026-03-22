# Deploy S3 bucket for Lambda source
deploy-bucket:
    aws cloudformation deploy \
        --template-file infra/bucket.yaml \
        --stack-name {{ project_name }}-bucket \
        --parameter-overrides BucketName=$SAVETOINK_BUCKET

# Deploy ACM certificate (must be deployed to us-east-1)
deploy-cert:
    @echo "Open https://us-east-1.console.aws.amazon.com/acm/certificates/ and add DNS validation records for $SAVETOINK_DOMAIN"
    aws cloudformation deploy \
        --template-file infra/cert.yaml \
        --stack-name {{ project_name }}-cert \
        --region us-east-1 \
        --parameter-overrides ProjectName={{ project_name }} DomainName="$SAVETOINK_DOMAIN"

# Deploy Lambda infrastructure
deploy-api: lint-cfn
    aws cloudformation deploy \
        --template-file infra/api.yaml \
        --stack-name {{ project_name }}-infra \
        --capabilities CAPABILITY_NAMED_IAM \
        --parameter-overrides \
            APIKeySecret="$SAVETOINK_API_KEY" \
            AppURL="$SAVETOINK_APP_URL" \
            Auth0Audience="$SAVETOINK_AUTH0_AUDIENCE" \
            Auth0ClientId="$SAVETOINK_AUTH0_CLIENT_ID" \
            Auth0ClientSecret="$SAVETOINK_AUTH0_CLIENT_SECRET" \
            Auth0Domain="$SAVETOINK_AUTH0_DOMAIN" \
            AuthBackend="${SAVETOINK_AUTH_BACKEND}" \
            BackupBucketName="$SAVETOINK_BACKUP_BUCKET_NAME" \
            BrowserlessKey="${SAVETOINK_BROWSERLESS_KEY}" \
            CertificateArn=$(just get-cert-arn) \
            CORSAllowOrigin="${SAVETOINK_CORS_ALLOW_ORIGIN}" \
            Debug="${SAVETOINK_DEBUG}" \
            DestEmail="$SAVETOINK_DEST_EMAIL" \
            DisableQuotaCheck="${SAVETOINK_DISABLE_QUOTA_CHECK}" \
            DomainName="$SAVETOINK_DOMAIN" \
            GitHubEnvironmentName="${SAVETOINK_GITHUB_ENVIRONMENT_NAME}" \
            LoggingProvider=${SAVETOINK_LOGGING_PROVIDER} \
            MailBackend=${SAVETOINK_EMAIL_BACKEND} \
            MailjetAPIKey="$SAVETOINK_MAILJET_API_KEY" \
            MailjetAPISecret="$SAVETOINK_MAILJET_API_SECRET" \
            MailjetWebhookSecret=${SAVETOINK_MAILJET_WEBHOOK_SECRET} \
            ProcessorSourceBucketKey={{ lambda_processor_archive }} \
            ProjectName={{ project_name }} \
            SchedulerSourceBucketKey={{ lambda_scheduler_archive }} \
            SenderEmail="$SAVETOINK_SENDER_EMAIL" \
            SentryDSN=${SAVETOINK_SENTRY_DSN} \
            SentryEnvironment=${SAVETOINK_SENTRY_ENVIRONMENT} \
            SentrySampleRate=${SAVETOINK_SENTRY_SAMPLE_RATE} \
            SourceBucketKey={{ lambda_archive }} \
            SourceBucketName=${SAVETOINK_BUCKET}

# Upload API Lambda source zip to S3
upload-zip-api: build-lambda-zip-api
    aws s3 cp bin/{{ lambda_archive }} s3://${SAVETOINK_BUCKET}/{{ lambda_archive }}

# Build Lambda binary for Linux
build-lambda-api:
    GOOS=linux GOARCH=amd64 go build -o bin/bootstrap ./backend/cmd/lambda

# Build Lambda zip for deployment
[working-directory('bin')]
build-lambda-zip-api: build-lambda-api
    zip {{ lambda_archive }} bootstrap

# Update Lambda function code only
deploy-lambda-api: build-lambda-zip-api upload-zip-api
    aws lambda update-function-code \
        --function-name {{ project_name }}-api \
        --s3-bucket ${SAVETOINK_BUCKET} \
        --s3-key {{ lambda_archive }} \
        --publish

# Upload Processor Lambda source zip to S3
upload-zip-processor: build-lambda-zip-processor
    aws s3 cp bin/{{ lambda_processor_archive }} s3://${SAVETOINK_BUCKET}/{{ lambda_processor_archive }}

# Build Lambda binary for Linux
build-lambda-processor:
    GOOS=linux GOARCH=amd64 go build -o bin/bootstrap ./backend/cmd/lambda/processor

# Build Lambda zip for deployment
[working-directory('bin')]
build-lambda-zip-processor: build-lambda-processor
    zip {{ lambda_processor_archive }} bootstrap

# Update Lambda function code only
deploy-lambda-processor: build-lambda-zip-processor upload-zip-processor
    aws lambda update-function-code \
        --function-name {{ project_name }}-processor \
        --s3-bucket ${SAVETOINK_BUCKET} \
        --s3-key {{ lambda_processor_archive }} \
        --publish

# Upload Scheduler Lambda source zip to S3
upload-zip-scheduler: build-lambda-zip-scheduler
    aws s3 cp bin/{{ lambda_scheduler_archive }} s3://${SAVETOINK_BUCKET}/{{ lambda_scheduler_archive }}

# Build Lambda binary for Linux
build-lambda-scheduler:
    GOOS=linux GOARCH=amd64 go build -o bin/bootstrap ./backend/cmd/lambda/scheduler

# Build Lambda zip for deployment
[working-directory('bin')]
build-lambda-zip-scheduler: build-lambda-scheduler
    zip {{ lambda_scheduler_archive }} bootstrap

# Update Lambda function code only
deploy-lambda-scheduler: build-lambda-zip-scheduler upload-zip-scheduler
    aws lambda update-function-code \
        --function-name {{ project_name }}-scheduler \
        --s3-bucket ${SAVETOINK_BUCKET} \
        --s3-key {{ lambda_scheduler_archive }} \
        --publish
