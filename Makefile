NAME := branchcheck
ARCH := amd64
VERSION := 1.0
DATE := $(shell date)
COMMIT_ID := $(shell git rev-parse --short HEAD)
SDK_INFO := $(shell go version)
LD_FLAGS := -X main.version $(VERSION) -X main.commit $(COMMIT_ID) -X main.buildTime '$(DATE)' -X main.sdkInfo '$(SDK_INFO)'

all: 
	go clean
	go test -v
	go build -ldflags "$(LD_FLAGS)"
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LD_FLAGS)" -o branchcheck-linux

