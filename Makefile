generate:
	go generate ./...

mocks: generate

build-docker:
	docker-compose build

docker-compose:
	docker-compose up

test:
	go test -json ./...

run:
	go run ./cmd/feeder/main.go