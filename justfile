PACKAGE_NAME := "github.com/NibiruChain/pricefeeder"
GOLANG_CROSS_VERSION := "v1.19.4"

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

# Run all repo tests, including heavy ones. Ex: just test ./feeder/... -run TestFeeder
test *ARGS:
  #!/usr/bin/env bash
  echo "Note that you can pass args to augment the test command."
  echo "For example: just test ./feeder/... -run TestFeeder"
  echo "  becomes -> go test ./feeder/... -run TestFeeder 2>&1 | tee out.txt"
  echo -e "\nThis takes a ~4 minutes to run. The '.' getting printed signifies 2 seconds have passed."

  args="{{ARGS}}"
  args="${args:-./...}"
  echo -e "\nRunning: go test $args 2>&1 | tee out.txt"

  # Run go test in a new process group so we can kill all children on interrupt
  set -m  # Enable job control to create process groups
  (go test $args 2>&1 | tee out.txt) & 
  go_test_pid=$!
  pgid=$(ps -o pgid= -p $go_test_pid 2>/dev/null | tr -d ' ')
  # Set up trap to kill the entire process group on interrupt
  trap "kill -TERM -${pgid:-$go_test_pid} 2>/dev/null; wait $go_test_pid 2>/dev/null; exit 130" INT TERM
  # Print while the go test process ID (PID) is running
  while kill -0 $go_test_pid 2>/dev/null; do
    printf '.'
    sleep 2
  done
  wait $go_test_pid
  # Capture exit code before running echo. 
  # Otherwise, the subshell exits with 0, hiding the true result.
  exit_code=$?
  trap - INT TERM
  set +m
  echo
  echo "Done."
  exit $exit_code

# Run the main application
run:
    go run ./main.go

# Run the main application in debug mode
run-debug:
    go run ./main.go -debug true

# Build the application
build:
    #!/usr/bin/env bash
    VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "dev")
    COMMIT=$(git rev-parse HEAD)
    CGO_ENABLED=0 go build -mod=readonly -ldflags="-s -w -X github.com/NibiruChain/pricefeeder/cmd.Version=${VERSION} -X github.com/NibiruChain/pricefeeder/cmd.CommitHash=${COMMIT}" .

# Install the application
install:
    #!/usr/bin/env bash
    VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "dev")
    COMMIT=$(git rev-parse HEAD)
    CGO_ENABLED=0 go install -mod=readonly -ldflags="-s -w -X github.com/NibiruChain/pricefeeder/cmd.Version=${VERSION} -X github.com/NibiruChain/pricefeeder/cmd.CommitHash=${COMMIT}" .

# Alias for both build and install
build-install: build install

# Run golang formatter (gofumpt)
fmt:
  gofumpt -w .

# Run `golangci-lint` with docker
lint: 
  #!/usr/bin/env bash
  echo "Running golangci-lint with docker!"
  image_version="v2.6.1"
  docker run --rm \
    -v "$(pwd)":/app \
    -v ~/.cache/golangci-lint/$image_version:/root/.cache \
    -w /app \
    golangci/golangci-lint:$image_version \
    golangci-lint run -v --fix 2>&1
