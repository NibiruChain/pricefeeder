PACKAGE_NAME := "github.com/NibiruChain/pricefeeder"
GOLANG_CROSS_VERSION := "v1.19.4"
VERSION := `git describe --tags --abbrev=0`
COMMIT := `git rev-parse HEAD`

# Displays available recipes by running `just -l`.
setup:
  #!/usr/bin/env bash
  just -l

# Generate Go files
generate:
    go generate ./...

# Build Docker image
build-docker:
    docker-compose build

# Run Docker Compose
docker-compose:
    docker-compose up

# Run all repo tests, including heavy ones
test:
  #!/usr/bin/env bash
  echo "This takes a few minutes to run. The '.' getting printed signifies 2 seconds have passed."
  (go test ./... & go_test_pid=$!
    # Print while the go test process ID (PID) is running
    while kill -0 $go_test_pid 2>/dev/null; do
      printf '.'
      sleep 2
    done
    wait $go_test_pid
    # Capture exit code before running echo. 
    # Otherwise, the subshell exits with 0, hiding the true result.
    exit_code=$?
    echo
    echo "Done."
    exit $exit_code)

# Run the main application
run:
    go run ./main.go

# Run the main application in debug mode
run-debug:
    go run ./main.go -debug true

# Build the application
build:
    CGO_ENABLED=0 go build -mod=readonly -ldflags="-s -w -X github.com/NibiruChain/pricefeeder/cmd.Version={{VERSION}} -X github.com/NibiruChain/pricefeeder/cmd.CommitHash={{COMMIT}}" .

# Install the application
install:
    CGO_ENABLED=0 go install -mod=readonly -ldflags="-s -w -X github.com/NibiruChain/pricefeeder/cmd.Version={{VERSION}} -X github.com/NibiruChain/pricefeeder/cmd.CommitHash={{COMMIT}}" .

# Alias for both build and install
build-install: build install

# Runs golang formatter (gofumpt)
fmt:
  gofumpt -w .
