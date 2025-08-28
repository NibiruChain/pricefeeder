# ==============================================================================
# Project metadata
# ==============================================================================
BRANCH  := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT  := $(shell git log -1 --format='%H')
VERSION ?= $(shell git describe --tags --abbrev=0)

ifeq ($(strip $(VERSION)),)
  VERSION := $(BRANCH)-$(COMMIT)
endif

# ==============================================================================
# Platform detection
# ==============================================================================
OS_NAME  := $(shell uname -s | tr A-Z a-z)

ifeq ($(OS_NAME),darwin)
  ARCH_NAME := all
else ifeq ($(shell uname -m),x86_64)
  ARCH_NAME := amd64
else
  ARCH_NAME := arm64
endif

# ==============================================================================
# Directories and dependencies
# ==============================================================================
WASMVM_VERSION := $(shell go list -m github.com/CosmWasm/wasmvm | awk '{sub(/^v/, "", $$2); print $$2}')

BUILDDIR ?= $(CURDIR)/build
TEMPDIR  ?= $(CURDIR)/temp

# ==============================================================================
# Build tags
# ==============================================================================
build_tags = netgo osusergo static

ifeq ($(OS_NAME),darwin)
  build_tags += static_wasm
else
  build_tags += muslc
endif

build_tags := $(strip $(build_tags))

# ==============================================================================
# Flags
# ==============================================================================
ldflags = -X github.com/NibiruChain/pricefeeder/cmd.Version=$(VERSION) \
          -X github.com/NibiruChain/pricefeeder/cmd.CommitHash=$(COMMIT) \
          -linkmode=external -w -s

ldflags    := $(strip $(ldflags))
BUILD_FLAGS = -tags "$(build_tags)" -ldflags '$(ldflags)'

CGO_LDFLAGS := -L$(TEMPDIR)/wasmvm/$(WASMVM_VERSION)/lib/$(OS_NAME)_$(ARCH_NAME)/
ifeq ($(OS_NAME),linux)
  CGO_LDFLAGS += -static
endif

# ==============================================================================
# Targets
# ==============================================================================

.PHONY: build install generate docker-build docker-run test run run-debug

$(TEMPDIR)/:
	mkdir -p $@

$(BUILDDIR)/:
	mkdir -p $@

# ------------------------------------------------------------------------------

# Download wasmvm static lib if missing
wasmvmlib: $(TEMPDIR)/
	@mkdir -p $(TEMPDIR)/wasmvm/$(WASMVM_VERSION)/lib/$(OS_NAME)_$(ARCH_NAME)/
	@if [ ! -f $(TEMPDIR)/wasmvm/$(WASMVM_VERSION)/lib/$(OS_NAME)_$(ARCH_NAME)/libwasmvm*.a ]; then \
	  if [ "$(OS_NAME)" = "darwin" ]; then \
	    wget https://github.com/CosmWasm/wasmvm/releases/download/v$(WASMVM_VERSION)/libwasmvmstatic_darwin.a \
	         -O $(TEMPDIR)/wasmvm/$(WASMVM_VERSION)/lib/$(OS_NAME)_$(ARCH_NAME)/libwasmvmstatic_darwin.a; \
	  else \
	    if [ "$(ARCH_NAME)" = "amd64" ]; then \
	      wget https://github.com/CosmWasm/wasmvm/releases/download/v$(WASMVM_VERSION)/libwasmvm_muslc.x86_64.a \
	           -O $(TEMPDIR)/wasmvm/$(WASMVM_VERSION)/lib/$(OS_NAME)_$(ARCH_NAME)/libwasmvm_muslc.a; \
	    else \
	      wget https://github.com/CosmWasm/wasmvm/releases/download/v$(WASMVM_VERSION)/libwasmvm_muslc.aarch64.a \
	           -O $(TEMPDIR)/wasmvm/$(WASMVM_VERSION)/lib/$(OS_NAME)_$(ARCH_NAME)/libwasmvm_muslc.a; \
	    fi; \
	  fi; \
	fi

# ------------------------------------------------------------------------------

# Build targets
build: BUILDARGS=-o $(BUILDDIR)/
build install: $(BUILDDIR)/ wasmvmlib
	CGO_ENABLED=1 CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" \
		go $@ -mod=readonly -trimpath $(BUILD_FLAGS) $(BUILDARGS)

# ------------------------------------------------------------------------------

# Utilities
generate:
	go generate ./...

docker-build:
	@docker build -t pricefeeder .

.env:
	@touch .env

docker-run: .env
	@docker run -it --rm --env-file .env pricefeeder $(CMD)

test:
	go test ./...

run:
	go run ./main.go

run-debug:
	go run ./main.go -debug true