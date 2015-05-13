NAME := branchcheck
ARCH := amd64
VERSION := 1.2
DATE := $(shell date)
COMMIT_ID := $(shell git rev-parse --short HEAD)
SDK_INFO := $(shell go version)
LD_FLAGS := -X main.buildInfo 'Version: $(VERSION), commitID: $(COMMIT_ID), build date: $(DATE), SDK: $(SDK_INFO)'

all: clean
	go test -v
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LD_FLAGS)" -o branchcheck-darwin-amd64
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LD_FLAGS)" -o branchcheck-linux-amd64
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LD_FLAGS)" -o branchcheck-windows-amd64

package: all
	mkdir -p packaging
	cp branchcheck-linux-amd64 packaging/$(NAME)
	fpm -s dir -t deb -v $(VERSION) -n $(NAME) -a amd64  -m"Mark Petrovic <mark.petrovic@xoom.com>" --url https://github.com/xoom/branchcheck --iteration 1 --prefix /usr/local/bin -C packaging .
	fpm -s dir -t rpm --rpm-os linux -v $(VERSION) -n $(NAME) -a amd64  -m"Mark Petrovic <mark.petrovic@xoom.com>" --url https://github.com/xoom/branchcheck --iteration 1 --prefix /usr/local/bin -C packaging .

clean:
	rm -rf *.deb *.rpm packaging
	go clean
	rm -f branchcheck*amd64
