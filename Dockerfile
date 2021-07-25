FROM golang:1.16-buster AS build
WORKDIR /go/src/go-diff
ENV CGO_ENABLED=0
ENV GO111MODULE=on 
RUN go install github.com/golang/mock/mockgen@v1.6.0
COPY . .
RUN go generate ./... && go test ./... && go install

FROM alpine:3
WORKDIR /opt/go-diff
COPY --from=build /go/bin/go-diff bin/go-diff
RUN chmod +x bin/go-diff
ENV GIN_MODE=release
ENTRYPOINT ["bin/go-diff"] 
