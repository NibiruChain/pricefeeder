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