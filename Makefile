NAME := branchcheck
ARCH := amd64
VERSION := 1.2
DATE := $(shell date)
COMMIT_ID := $(shell git rev-parse --short HEAD)
SDK_INFO := $(shell go version)
LD_FLAGS := -X main.buildInfo 'Version: $(VERSION), commitID: $(COMMIT_ID), build date: $(DATE), SDK: $(SDK_INFO)'

all: 
	rm -f branchcheck-linux
	go clean
	go test -v
	go build -ldflags "$(LD_FLAGS)"
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LD_FLAGS)" -o branchcheck-linux

package: all
	rm -f *.deb *.rpm
	rm -rf  packaging
	mkdir -p packaging
	cp branchcheck-linux packaging/branchcheck
	fpm -s dir -t deb -v $(VERSION) -n $(NAME) -a amd64  -m"Mark Petrovic <mark.petrovic@xoom.com>" --url https://github.com/xoom/branchcheck --iteration 1 --prefix /usr/local/bin -C packaging .
	fpm -s dir -t rpm --rpm-os linux -v $(VERSION) -n $(NAME) -a amd64  -m"Mark Petrovic <mark.petrovic@xoom.com>" --url https://github.com/xoom/branchcheck --iteration 1 --prefix /usr/local/bin -C packaging .

