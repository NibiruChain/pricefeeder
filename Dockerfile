# --------- Stage: Build pricefeeder
FROM golang:1.24 AS builder

WORKDIR /feeder

RUN apt-get update && apt-get install -y --no-install-recommends \
    liblz4-dev libsnappy-dev zlib1g-dev libbz2-dev libzstd-dev

COPY go.sum go.mod ./
RUN go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/go/pkg \
  make build

# --------- Stage: Run the binary
FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /feeder/pricefeeder .
USER nonroot:nonroot
ENTRYPOINT ["/pricefeeder"]
