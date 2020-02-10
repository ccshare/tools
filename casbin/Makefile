BINARY=myauth
BUILDDATE=$(shell date +'%Y-%m-%dT%H:%M:%SZ')
VERSION=1.0.0
LONGVER=${VERSION}@${BUILDDATE}@$(shell git log --format=%h -n 1)

LDFLAGS=-ldflags "-X main.version=${LONGVER}"

.DEFAULT_GOAL:=default

test:
	go test ./...

vet:
	go vet ./...

default:
	@echo "Building ${BINARY}-${VERSION}"
	go build ${LDFLAGS} -o ${BINARY}

install: default
	install ${BINARY} /usr/local/bin/

clean:
	rm -rf *zip
	rm -rf ${BINARY}

.PHONY: test vet default clean
