FROM golang:alpine AS builder

RUN apk add --no-cache build-base git

WORKDIR /feeder

COPY go.sum go.mod ./
RUN go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/go/pkg \
  make build

FROM gcr.io/distroless/static:nonroot

WORKDIR /
COPY --from=builder /feeder/pricefeeder .
USER nonroot:nonroot
ENTRYPOINT ["/pricefeeder"]
