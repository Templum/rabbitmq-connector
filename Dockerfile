FROM golang:1.10

RUN mkdir -p /go/src/github.com/Templum/openfaas-rabbitmq-connector/

WORKDIR /go/src/github.com/Templum/openfaas-rabbitmq-connector

COPY . .

RUN gofmt -l -d $(find . -type f -name '*.go' -not -path "./vendor/*") && \
  VERSION=$(git describe --all --exact-match `git rev-parse HEAD` | grep tags | sed 's/tags\///') && \
  GIT_COMMIT=$(git rev-list -1 HEAD) && \
  CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w \
  -X github.com/Templum/openfaas-rabbitmq-connector/pkg/version.Version=${VERSION} \
  -X github.com/Templum/openfaas-rabbitmq-connector/pkg/version.GitCommit=${GIT_COMMIT}" \
  -a -installsuffix cgo -o rmq-connector .

FROM alpine:3.7

RUN addgroup -S app \
    && adduser -S -g app app \
    && apk --no-cache add ca-certificates

WORKDIR /home/app

COPY --from=0 /go/src/github.com/Templum/openfaas-rabbitmq-connector/rmq-connector .

RUN chown -R app:app ./

USER app

ENTRYPOINT ["./rmq-connector"]