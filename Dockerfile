ARG src=base

# ---------- Build Stage ----------
FROM golang:1.24 AS build-base

WORKDIR /feeder

COPY go.sum go.mod ./
RUN go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg \
    --mount=type=cache,target=/feeder/temp \
  make build && cp build/pricefeeder /root/

# ---------- Binary Copy (External Build) ----------
FROM busybox AS build-external

WORKDIR /root
COPY ./dist/ /root/

ARG TARGETARCH
RUN if [ "${TARGETARCH}" = "arm64" ]; then \
      cp arm64/pricefeeder /root/pricefeeder; \
    else \
      cp amd64/pricefeeder /root/pricefeeder; \
    fi

# ---------- Binary Build Source ----------
FROM build-${src} AS build-source

FROM gcr.io/distroless/static:nonroot

COPY --from=build-source /root/pricefeeder /bin/
USER nonroot:nonroot
ENTRYPOINT ["/bin/pricefeeder"]
