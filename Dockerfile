FROM golang:1.19

ARG ARCH="aarch64"

WORKDIR /feeder
COPY . .

RUN go build -o ./build/feeder ./cmd/feeder/.
ENTRYPOINT ["./build/feeder"]
