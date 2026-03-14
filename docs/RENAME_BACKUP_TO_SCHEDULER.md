# Rename "backup" Lambda to "scheduler"

## Overview
The backup lambda is being renamed to "scheduler" to reflect its more generic purpose of handling scheduled administrative tasks beyond just backups.

## Files to Rename

### 1. Backend Lambda Code
- **Directory/File:** `backend/cmd/lambda/backup/main.go` → `backend/cmd/lambda/scheduler/main.go`
- **Directory/File:** `backend/lib/backup/lambda/handler.go` → `backend/lib/scheduler/lambda/handler.go`

## Code Changes

### `backend/cmd/lambda/scheduler/main.go`
- Update import path:
  - From: `github.com/shaftoe/savetoink/backend/lib/backup/lambda`
  - To: `github.com/shaftoe/savetoink/backend/lib/scheduler/lambda`

### `backend/lib/scheduler/lambda/handler.go`
- Update comment on line 10:
  - From: `// HandleEvent is the entry point for the backup cronjob Lambda.`
  - To: `// HandleEvent is the entry point for the scheduler cronjob Lambda.`
- Update log message on line 12:
  - From: `slog.Info("backup cronjob triggered", "event", event)`
  - To: `slog.Info("scheduler cronjob triggered", "event", event)`

## Infrastructure Changes (infra/api.yaml)

### Parameters Section
- Line 21-23: Rename parameter
  - `BackupSourceBucketKey` → `SchedulerSourceBucketKey`
  - Update description: "S3 key for Backup Lambda source zip file" → "S3 key for Scheduler Lambda source zip file"

### IAM Roles Section
- Line 224-262: Rename role
  - `BackupLambdaExecutionRole` → `SchedulerLambdaExecutionRole`
  - Line 227: RoleName: `${ProjectName}-backup-lambda-exec` → `${ProjectName}-scheduler-lambda-exec`
  - Line 238: PolicyName: `${ProjectName}-backup-dynamodb-access` → `${ProjectName}-scheduler-dynamodb-access`
  - Line 251: PolicyName: `${ProjectName}-backup-s3-access` → `${ProjectName}-scheduler-s3-access`

### Lambda Function Section
- Line 393-420: Rename function
  - `BackupLambdaFunction` → `SchedulerLambdaFunction`
  - Line 396: FunctionName: `${ProjectName}-backup` → `${ProjectName}-scheduler`
  - Line 397: Description: "Backup - Daily DynamoDB backup to S3" → "Scheduler - Scheduled administrative tasks"
  - Line 403: Code.S3Key: `!Ref BackupSourceBucketKey` → `!Ref SchedulerSourceBucketKey`
  - Environment variables:
    - Keep `SAVETOINK_BACKUP_BUCKET_NAME` unchanged (referencing the S3 bucket which stays named "backups")

### Log Group Section
- Line 422-426: Rename log group
  - `BackupLambdaLogGroup` → `SchedulerLambdaLogGroup`
  - Line 425: LogGroupName reference to update

### Permission Section
- Line 428-434: Rename permission
  - `BackupLambdaPermission` → `SchedulerLambdaPermission`
  - Line 431: FunctionName reference to update
  - Line 433: SourceArn reference to update

### Schedule Rule Section
- Line 439-448: Rename schedule rule
  - `BackupScheduleRule` → `SchedulerScheduleRule`
  - Line 442: Name: `${ProjectName}-backup-schedule` → `${ProjectName}-scheduler-schedule`
  - Line 443: Description: "Daily backup of DynamoDB tables to S3" → "Daily execution of scheduled administrative tasks"
  - Line 447: Target: `BackupLambdaTarget` → `SchedulerLambdaTarget`

### S3 Bucket Section
- Line 650: Keep `BackupS3Bucket` unchanged
  - BucketName: `${ProjectName}-backups` stays as-is
- Line 824-834: Keep exports unchanged
  - `BackupS3BucketName` export stays as-is
  - `BackupS3BucketArn` export stays as-is

### Outputs Section
- Line 836-846: Rename function outputs
  - `BackupFunctionName` → `SchedulerFunctionName`
  - `BackupFunctionArn` → `SchedulerFunctionArn`
  - Line 839-840: Export names
    - `${ProjectName}-backup-function-name` → `${ProjectName}-scheduler-function-name`
    - `${ProjectName}-backup-function-arn` → `${ProjectName}-scheduler-function-arn`
- Line 848-851: Rename schedule rule output
  - `BackupScheduleRuleArn` → `SchedulerScheduleRuleArn`
  - Line 850: Export name
    - `${ProjectName}-backup-schedule-rule-arn` → `${ProjectName}-scheduler-schedule-rule-arn`

## New Build/Deploy Commands (.justfiles/infra.just)

Add the following commands after line 91, following the pattern of api/processor lambdas:

```just
# Upload Scheduler Lambda source zip to S3
upload-zip-scheduler:
    aws s3 cp bin/{{ lambda_scheduler_archive }} s3://${SAVETOINK_BUCKET}/{{ lambda_scheduler_archive }}

# Build Scheduler Lambda binary for Linux
build-lambda-scheduler:
    GOOS=linux GOARCH=amd64 go build -o bin/bootstrap ./backend/cmd/lambda/scheduler

# Build Scheduler Lambda zip for deployment
[working-directory('bin')]
build-lambda-zip-scheduler: build-lambda-scheduler
    zip {{ lambda_scheduler_archive }} bootstrap

# Update Scheduler Lambda function code only
deploy-lambda-scheduler: build-lambda-zip-scheduler upload-zip-scheduler
    aws lambda update-function-code \
        --function-name {{ project_name }}-scheduler \
        --s3-bucket ${SAVETOINK_BUCKET} \
        --s3-key {{ lambda_scheduler_archive }} \
        --publish
```

Update the justfile root to include the new archive variable:
- Add after line 5: `lambda_scheduler_archive := 'lambda-scheduler-source.zip'`

Update the deploy command in `.justfiles/deploy.just`:
- Add after line 8: `just build-lambda-zip-scheduler`
- Add after line 9: `just upload-zip-scheduler`

Update GitHub workflows:
- `.github/workflows/deploy-lambda.yml`: Add scheduler deployment step

## Deployment Notes

### Breaking Changes
- Lambda function name changes: `<project-name>-backup` → `<project-name>-scheduler`
- CloudFormation outputs will change names
- EventBridge schedule rule ARN will change

### Migration Steps
1. Rename code files and directories
2. Update code imports and references
3. Update infra/api.yaml with new resource names
4. Add build/deploy commands to justfiles
5. Update GitHub workflows
6. Deploy the new scheduler lambda infrastructure
7. Verify the schedule rule is triggering the new lambda
8. Monitor CloudWatch logs for the new scheduler lambda

### Verification Checklist
- [ ] Code files renamed and updated
- [ ] Infrastructure template updated
- [ ] Build commands added
- [ ] GitHub workflows updated
- [ ] Lambda function deployed successfully
- [ ] Schedule rule configured
- [ ] CloudWatch logs show scheduler execution
- [ ] Backup functionality still works (if implemented)

## Rollback Plan
If issues occur:
1. Delete the scheduler lambda deployment
2. Restore the original backup lambda infrastructure from git
3. Deploy the backup lambda again
4. Verify functionality is restored
