set dotenv-load := true

project_name := 'savetoink'
lambda_archive := 'lambda-source.zip'
lambda_processor_archive := 'lambda-processor-source.zip'
lambda_scheduler_archive := 'lambda-scheduler-source.zip'
build_flags := "-X github.com/shaftoe/savetoink/backend/lib/consts.version"

import '.justfiles/auth0.just'
import '.justfiles/aws.just'
import '.justfiles/build.just'
import '.justfiles/deploy.just'
import '.justfiles/deps.just'
import '.justfiles/dev.just'
import '.justfiles/dynamodb.just'
import '.justfiles/extension.just'
import '.justfiles/infra.just'
import '.justfiles/mail.just'
import '.justfiles/misc.just'
import '.justfiles/smoketest.just'
import '.justfiles/test.just'
import '.justfiles/website.just'
import '.justfiles/frontend.just'
