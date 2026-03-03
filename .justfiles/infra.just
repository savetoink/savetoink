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

# Upload Lambda source zip to S3
upload-zip:
    aws s3 cp bin/{{ lambda_archive }} s3://${SAVETOINK_BUCKET}/{{ lambda_archive }}

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
            Auth0ClientId="$SAVETOINK_AUTH0_CLIENT_ID" \
            Auth0ClientSecret="$SAVETOINK_AUTH0_CLIENT_SECRET" \
            Auth0Domain="$SAVETOINK_AUTH0_DOMAIN" \
            AuthBackend="auth0" \
            CertificateArn=$(just get-cert-arn) \
            DestEmail="$SAVETOINK_DEST_EMAIL" \
            DomainName="$SAVETOINK_DOMAIN" \
            MailBackend=${SAVETOINK_EMAIL_BACKEND} \
            MailjetAPIKey="$SAVETOINK_MAILJET_API_KEY" \
            MailjetAPISecret="$SAVETOINK_MAILJET_API_SECRET" \
            MailjetWebhookSecret=${SAVETOINK_MAILJET_WEBHOOK_SECRET} \
            ProjectName={{ project_name }} \
            SenderEmail="$SAVETOINK_SENDER_EMAIL" \
            SourceBucketKey={{ lambda_archive }} \
            SourceBucketName=${SAVETOINK_BUCKET} \
            Debug="true"

# Update Lambda function code only
deploy-lambda: build-lambda-zip upload-zip
    aws lambda update-function-code \
        --function-name {{ project_name }}-api \
        --s3-bucket ${SAVETOINK_BUCKET} \
        --s3-key {{ lambda_archive }} \
        --publish
