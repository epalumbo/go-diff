
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

Using Docker (recommended):
```sh
$ docker build -t go-diff .
```

# Running the API server
During development the service can be run locally as follows:
```sh
$ go run .
```
For production use, after build, run the generated executable file.
In any case, do not forget to configure the AWS SDK (credentials and region via local profile or environment variables, or execution roles when running in the cloud) and the bucket name. See docker-compose.yml for required variables.

Using Docker Compose (recommended):
```sh
$ docker compose up
```
* HTTP service will listen on local port 8080.
* You can easily define environment variables by placing an untracked ".env" file in the root directory.

To stop the services and clean:
```sh
$ docker compose down
```

To run the service container using Docker without Compose:
```sh
$ docker run -p 8080:8080 go-diff
```
