
# Prerequisites

* Go 1.16+
* GoMock 1.6.0
    * Install GoMock: `go install github.com/golang/mock/mockgen@v1.6.0`

* Optional: Docker. There is a multi-stage build available, no need to install any of the previously mentioned dependencies when no development work is required.

# Build
Binary can be quickly built locally as follows:
```sh
$ go generate ./... && go build
```

With test execution:
```sh
$ go generate ./... && go test ./... && go build
```

Using Docker (recommended, required for deployment):
```sh
$ docker build -t go-diff .
```

# Deploying to AWS

### Manual deployment
The application shall be deployed to AWS Lambda as a Docker container.
- After building the Docker image, it has to be pushed to a private AWS ECR repository.
- Then it can be configured as container for AWS Lambda.
- The Lambda application requires "AWS_BUCKET_NAME" as environment variable, pointing to a S3 bucket in the same region.
- The Lambda execution role has to allow RW access to the provided S3 bucket.
- An API Gateway proxy integration must be configured to invoke the deployed Lambda function.

### Deploy via SAM (recommended)
The application can also be deployed to AWS using the [SAM](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/index.html) CLI. 
A SAM template is available in the root directory of the project.
```bash
sam build && sam deploy --guided
```
- `sam build`: locally builds the application as a Docker image and prepares it for deployment.
- `sam deploy --guided`: deploys the application stack defined in the SAM template to AWS, 
interactively asking for parameters and configuration.
  - This requires an ECR private repository to be specified.