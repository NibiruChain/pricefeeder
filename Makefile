generate:
	go generate ./...

mocks: generate

image:
	docker build -t nibiru/price-feeder:master .