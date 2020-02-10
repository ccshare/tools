BINARY=authd
BUILDDATE=$(shell date +'%Y-%m-%dT%H:%M:%SZ')@$(shell git rev-parse --short HEAD)
VERSION=1.0.$(shell git rev-list HEAD --count)
LONGVER=${VERSION}@${BUILDDATE}

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
	rm -rf ${BINARY}

.PHONY: pkg test vet default clean
