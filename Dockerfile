FROM golang:1.19.3-alpine AS builder

ARG NIBID_REPO="https://github.com/NibiruChain/nibiru"
ARG NIBID_COMMIT="961796e2263c2af44d83eeb7644eaa3308f55dd6"
ARG ARCH="aarch64"

RUN apk update && apk upgrade && \
    apk add --no-cache make bash git openssh build-base

WORKDIR /node

ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.1.1/libwasmvm_muslc.aarch64.a /lib/libwasmvm_muslc.aarch64.a
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.1.1/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.x86_64.a
RUN sha256sum /lib/libwasmvm_muslc.aarch64.a | grep 9ecb037336bd56076573dc18c26631a9d2099a7f2b40dc04b6cae31ffb4c8f9a
RUN sha256sum /lib/libwasmvm_muslc.x86_64.a | grep 6e4de7ba9bad4ae9679c7f9ecf7e283dd0160e71567c6a7be6ae47c81ebe7f32

RUN cp /lib/libwasmvm_muslc.${ARCH}.a /lib/libwasmvm_muslc.a

RUN git clone $NIBID_REPO .
RUN git checkout $NIBID_COMMIT
RUN LINK_STATICALLY=true BUILD_TAGS=muslc make build

WORKDIR /feeder

COPY . .
RUN go build -ldflags='-linkmode=external -extldflags "-Wl,-z,muldefs -static"' -tags=muslc -o ./build/feeder ./cmd/feeder/.

FROM alpine

ARG ARCH="aarch64"

COPY --from=builder /node/build/nibid /usr/bin/nibid
COPY --from=builder /feeder/build/feeder /usr/bin/feeder
COPY --from=builder /feeder/config/config.yaml /root/config.yaml
COPY ./scripts /scripts