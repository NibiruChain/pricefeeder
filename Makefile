generate:
	go generate ./...

mocks: generate

build-feeder:
	docker-compose build

run:
	docker-compose up price_feeder
