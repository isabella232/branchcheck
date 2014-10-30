LD_FLAGS := -X main.commit $(shell git rev-parse --short HEAD)

all: 
	go clean
	godep go test
	godep go build -ldflags "$(LD_FLAGS)"
