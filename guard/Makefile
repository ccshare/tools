BINARY=guard
BUILDDATE=$(shell date +'%Y-%m-%dT%H:%M:%SZ')
VERSION=1.0.0
LONGVER=${VERSION}@${BUILDDATE}

epochTime=2020-01-01T00:00:00Z
cloudgatewayService=pm2-root.service

LDFLAGS=-ldflags "-X main.version=${LONGVER} -X main.epochTime=${epochTime} -X main.cloudgatewayService=${cloudgatewayService}"

.DEFAULT_GOAL:=default
pkg:
	@echo "Building Linux amd64 ${BINARY}-${VERSION}"
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS}
	zip -m ${BINARY}-${VERSION}-linux.zip ${BINARY}
	
default:
	@echo "Building ${BINARY}-${VERSION}"
	go build ${LDFLAGS}

clean:
	rm -rf *zip
	rm -rf ${BINARY}

.PHONY: pkg default clean
