
# Prerequisites

* Go 1.15+
* GoMock 1.4+
    * Install GoMock: `GO111MODULE=on go get github.com/golang/mock/mockgen@v1.4.4`

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
For development purposes:
```sh
$ go run .
```

For production use, after build, just run the generated executable file.

Using Docker (recommended):
```sh
$ docker run -p 8080:8080 go-diff
```