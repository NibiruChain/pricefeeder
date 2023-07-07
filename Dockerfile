FROM golang:1.19 AS builder

WORKDIR /feeder

COPY go.sum go.mod ./
RUN go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/go/pkg \
  go build -o ./build/feeder ./cmd/feeder/.

FROM alpine:latest
WORKDIR /root

COPY --from=builder /feeder/build/feeder /usr/local/bin/feeder

ENTRYPOINT ["feeder"]
