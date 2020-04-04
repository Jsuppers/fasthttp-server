PROJECT_NAME := "fasthttp_server"

.PHONY: build lint

all: build

generate:
	@go get -v -d ./...
	go generate ./...

build:
	CGO_ENABLED=0 go build -o ./bin/${PROJECT_NAME} .

lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.22.2
	bin/golangci-lint run

coverage: generate
	go get golang.org/x/tools/cmd/cover
	go get github.com/mattn/goveralls
	go test -coverprofile=cover.out ./...
	go tool cover -func=cover.out
	rm cover.out
