
# Prerequisites

* Go 1.16+
* GoMock 1.4+
    * Install GoMock: `GO111MODULE=on go get github.com/golang/mock/mockgen@v1.4.4`
* Redis 6.x: persistent application storage, available via Docker. 

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
In any case, do not forget to set the `REDIS_URL` environment variable with the proper Redis URL connection string.

Using Docker Compose (recommended):
```sh
$ docker-compose up
```
* This will run the service container and a Redis database. 
* HTTP service will listen on local port 8080.

To stop the services and clean:
```sh
$ docker-compose down
```

Just to run the service container using Docker:
```sh
$ docker run -p 8080:8080 go-diff
```
