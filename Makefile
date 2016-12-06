pkgs = $(shell go list ./... | grep -v /vendor/)

# This how we want to name the binary output
BINARY=ebs-volumes

# These are the values we want to pass for VERSION and BUILD
VERSION=0.0.1
BUILD=`git rev-parse HEAD`

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.build=${BUILD}"

all: format build test

# Builds the project
build:
	go build ${LDFLAGS} -o ${BINARY}

# Installs our project: copies binaries
install:
	go install ${LDFLAGS}

# Cleans our project: deletes binaries
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

tools:
	go get -u github.com/kardianos/govendor  
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install

format:
	@echo ">> formatting code"
	go fmt $(pkgs)

test:
	@echo ">> running short tests"
	go test -short $(pkgs)

vet:
	@echo ">> vetting code"
	go vet $(pkgs)

lint:
	@echo ">> linting code"
	gometalinter --vendor ./...

.PHONY: clean install tools format vet lint