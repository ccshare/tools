APP?=gwload
BUILDDATE=$(shell date +'%Y-%m-%dT%H:%M:%SZ')
VERSION=1.0.4
LONGVER=${VERSION}@${BUILDDATE}@$(shell git rev-parse --short HEAD)

LDFLAGS=-ldflags "-X main.version=${LONGVER}"

.DEFAULT_GOAL:=default
## pkg: build and package the application
pkg:
	@echo "Building Linux amd64 ${APP}-${VERSION}"
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS}
	zip -m ${APP}-${VERSION}-linux.zip ${APP}
	
	@echo "Building Macos amd64 ${APP}-${VERSION}"
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS}
	zip -m ${APP}-${VERSION}-macos.zip ${APP}
	
	@echo "Building Windows amd64 ${APP}-${VERSION}"
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS}
	zip -m ${APP}-${VERSION}-win.zip ${APP}.exe
## test: runs go test with default values
test:
	go test ./...
## vet: runs go vet
vet:
	go vet ./...
## default: build the application
default:
	@echo "Building ${APP}-${VERSION}"
	go build ${LDFLAGS}
## install: install the application to /usr/local/bin/
install: default
	install ${APP} /usr/local/bin/
## clean: cleans the binary
clean:
	rm -rf *zip ${APP}
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: pkg test vet default clean help
