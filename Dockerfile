FROM golang:alpine AS builder

WORKDIR /feeder

COPY go.sum go.mod ./
RUN go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/go/pkg \
  go build -o ./build/feeder .

FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /feeder/build/feeder .
USER nonroot:nonroot
ENTRYPOINT ["/feeder"]
