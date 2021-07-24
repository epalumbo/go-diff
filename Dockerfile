FROM golang:1.16-buster AS build
WORKDIR /go/src/go-diff
ENV CGO_ENABLED=0
ENV GO111MODULE=on 
RUN go get github.com/golang/mock/mockgen@v1.4.4
COPY . .
RUN go generate ./... && go test ./... && go install

FROM alpine:3
WORKDIR /opt/go-diff
COPY --from=build /go/bin/go-diff bin/go-diff
RUN chmod +x bin/go-diff
ENV GIN_MODE=release
ENTRYPOINT ["bin/go-diff"] 
