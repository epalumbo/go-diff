
# Prerequisites

* Go 1.16+
* GoMock 1.6.0
    * Install GoMock: `go install github.com/golang/mock/mockgen@v1.6.0`

* Optional: Docker. There is a multi-stage build and a docker-compose file available, no need to install any of the previously mentioned dependencies when no development work is required.

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
The application shall be deployed to AWS Lambda as a Docker container.
- After building the Docker image, it has to be pushed to a private AWS ECR repository.
- Then it can be configured as container for AWS Lambda.
- The Lambda application requires "AWS_BUCKET_NAME" as environment variable, pointing to a S3 bucket in the same region.
- The Lambda execution role has to allow RW access to the provided S3 bucket.
- An API Gateway proxy integration must be configured to invoke the deployed Lambda function.
