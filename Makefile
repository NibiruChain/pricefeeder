generate:
	go generate ./...

mocks: generate

compose-all: image
	docker-compose down --volumes
	docker-compose up --build

image:
	docker build -t nibiru/price-feeder:master .