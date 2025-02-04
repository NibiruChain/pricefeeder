PACKAGE_NAME		  := github.com/NibiruChain/pricefeeder
GOLANG_CROSS_VERSION  ?= v1.19.4
VERSION ?= $(shell git describe --tags --abbrev=0)
COMMIT ?= $(shell git rev-parse HEAD)
BUILD_TARGETS := build install

generate:
	go generate ./...

build-docker:
	docker-compose build

docker-compose:
	docker-compose up

test:
	go test ./...

run:
	go run ./main.go

run-debug:
	go run ./main.go -debug true

###############################################################################
###                                Build                                    ###
###############################################################################

.PHONY: build install
$(BUILD_TARGETS):
	CGO_ENABLED=0 go $@ -mod=readonly -ldflags="-s -w -X github.com/NibiruChain/pricefeeder/cmd.Version=$(VERSION) -X github.com/NibiruChain/pricefeeder/cmd.CommitHash=$(COMMIT)" .
