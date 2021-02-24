FROM golang:1.15.8-alpine as base_builder

RUN apk --no-cache add ca-certificates git

WORKDIR /go/src/github.com/Templum/rabbitmq-connector/
ENV GO111MODULE=on

COPY go.mod go.sum  ./
RUN go mod download

FROM base_builder as builder
COPY . .

RUN VERSION=$(git describe --all --exact-match $(git rev-parse HEAD) | grep tags | sed 's/tags\///') && \
  GIT_COMMIT=$(git describe --always) && \
  echo "Git TAG: $VERSION GIT Commit: $GIT_COMMIT" && \
  CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w \
  -X github.com/Templum/rabbitmq-connector/pkg/version.Version=${VERSION} \
  -X github.com/Templum/rabbitmq-connector/pkg/version.GitCommit=${GIT_COMMIT}" \
  -a -installsuffix cgo -o rmq-connector .

FROM alpine:3.13.2

RUN addgroup -S app \
  && adduser -S -g app app \
  && apk --no-cache add ca-certificates

WORKDIR /home/app

COPY --from=builder /go/src/github.com/Templum/rabbitmq-connector/rmq-connector .

RUN chown -R app:app ./

USER app

ENTRYPOINT ["./rmq-connector"]