.PHONY: dev build generate install image release profile bench test clean setup

CGO_ENABLED=0
VERSION=$(shell git describe --abbrev=0 --tags)
COMMIT=$(shell git rev-parse --short HEAD)
PACKAGE=$(shell go list)

all: dev

dev: build
	@./bin/calypso --version
	@./bin/calypsod --version

build: clean generate
	@go build \
		-o ./bin/calypso \
		-tags "netgo static_build" -installsuffix netgo \
		-ldflags "-w -X $(PACKAGE)/internal.Version=$(VERSION) -X $(PACKAGE)/internal.Commit=$(COMMIT)" \
		./seagull/calypso/...
	@go build \
		-o ./bin/calypsod \
		-tags "netgo static_build" -installsuffix netgo \
		-ldflags "-w -X $(PACKAGE)/internal.Version=$(VERSION) -X $(PACKAGE)/internal.Commit=$(COMMIT)" \
		./seagull/calypsod/...

generate:
	@go generate $(PACKAGE)/...

install: build
	@go install ./seagull/calypso/...
	@go install ./seagull/calypsod/...

clean:
	@rm -rf ./bin/*
	@git clean -f -d -X
