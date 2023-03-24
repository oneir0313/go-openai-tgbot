FROM golang:1.19-alpine AS builder

WORKDIR /code

ENV CGO_ENABLED 0
ENV GOPATH /go
ENV GOCACHE /go-build

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/cache \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod/cache \
    --mount=type=cache,target=/go-build \
    go build -o bin/openai-tgbot main.go

CMD ["/code/bin/openai-tgbot"]

FROM alpine
COPY --from=builder /code/bin/openai-tgbot /
CMD ["/openai-tgbot"]