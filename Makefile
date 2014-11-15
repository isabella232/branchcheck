LD_FLAGS := -X main.commit $(shell git rev-parse --short HEAD)

all: deps
	go clean
	godep go test
	godep go build -ldflags "$(LD_FLAGS)"

deps:
	which godep || go get github.com/tools/godep
