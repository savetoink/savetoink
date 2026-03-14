# Full deployment (bucket + upload + infra)
deploy:
    just deploy-bucket
    just deploy-cert
    just upload-zip-api
    just upload-zip-processor
    just upload-zip-scheduler
    just deploy-api
    @echo "Add DNS record: $SAVETOINK_DOMAIN CNAME $(just get-distribution-url)."

# Destroy Lambda infrastructure
destroy:
    aws cloudformation delete-stack --stack-name {{ project_name }}-infra
    aws cloudformation wait stack-delete-complete --stack-name {{ project_name }}-infra
    -aws s3 rm s3://$SAVETOINK_BUCKET --recursive
    aws cloudformation delete-stack --stack-name {{ project_name }}-bucket
    aws cloudformation wait stack-delete-complete --stack-name {{ project_name }}-bucket
    aws cloudformation delete-stack --stack-name {{ project_name }}-cert --region us-east-1
    aws cloudformation wait stack-delete-complete --stack-name {{ project_name }}-cert --region us-east-1
