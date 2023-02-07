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

cycle: # remove from PR
	git push --delete origin v0.1.1-rc
	git tag -d v0.1.1-rc
	git tag v0.1.1-rc
	git push origin HEAD --tags