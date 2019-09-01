BINARY=logger
BUILDDATE=$(shell date +'%Y-%m-%dT%H:%M:%SZ')
VERSION=1.0.0
LONGVER=${VERSION}@${BUILDDATE}@$(shell git rev-list HEAD --count)

LDFLAGS=-ldflags "-X main.version=${LONGVER}"

.DEFAULT_GOAL:=default
test:
	go test ./...

vet:
	go vet ./...

default:
	@echo "Building ${BINARY}-${VERSION}"
	go build ${LDFLAGS}

install: default
	install ${BINARY} /usr/local/bin/

clean:
	go clean
	rm -rf *zip

.PHONY: test vet default clean
