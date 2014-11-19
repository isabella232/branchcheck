LD_FLAGS := -X main.commit $(shell git rev-parse --short HEAD)

all: deps
	go clean
	godep go test -v
	godep go build -ldflags "$(LD_FLAGS)"
	GOOS=linux GOARCH=amd64 godep go build -ldflags "$(LD_FLAGS)" -o branchcheck-linux

deps:
	which godep || go get github.com/tools/godep
