PROJECT_NAME := "fasthttp-server"

.PHONY: build lint

all: build

generate:
	@go get github.com/golang/mock/mockgen
	@go get -v -d ./...
	go generate ./...

build:
	CGO_ENABLED=0 go build -o ./bin/${PROJECT_NAME} .

lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.24.0
	bin/golangci-lint run

coverage: generate
	go get golang.org/x/tools/cmd/cover
	go get github.com/mattn/goveralls
	go test -coverprofile=cover.out ./...
	go tool cover -func=cover.out
	rm cover.out
