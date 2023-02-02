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
