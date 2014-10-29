LD_FLAGS := -X main.commit $(shell git rev-parse --short HEAD)

all: 
	go clean
	go test
	go build -ldflags "$(LD_FLAGS)"
