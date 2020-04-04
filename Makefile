PROJECT_NAME := "fasthttp_server"

.PHONY: build lint

all: build

build:
	CGO_ENABLED=0 go build -o ./bin/${PROJECT_NAME} .

lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.22.2
	bin/golangci-lint run
