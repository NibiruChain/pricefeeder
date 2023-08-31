generate:
	go generate ./...

build-docker:
	docker-compose build

docker-compose:
	docker-compose up

test:
	go test ./...

run:
	go run ./cmd/feeder/main.go

run-debug:
	go run ./cmd/feeder/main.go -debug true

###############################################################################
###                                Build                                    ###
###############################################################################

.PHONY: build
build:
	go build -mod=readonly ./...

.PHONY: install
install:
	go install -mod=readonly ./...

###############################################################################
###                               Release                                   ###
###############################################################################

PACKAGE_NAME		  := github.com/NibiruChain/pricefeeder
GOLANG_CROSS_VERSION  ?= v1.19.4

release:
	docker run \
		--rm \
		--platform=linux/amd64 \
		-v "$(CURDIR)":/go/src/$(PACKAGE_NAME) \
		-w /go/src/$(PACKAGE_NAME) \
		-e CGO_ENABLED=1 \
		-e GITHUB_TOKEN=${GITHUB_TOKEN} \
		goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		release --rm-dist
