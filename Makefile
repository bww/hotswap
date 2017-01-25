
export GOPATH := $(GOPATH):$(PWD)

TEST_PKGS := github.com/bww/go-hotswap

.PHONY: all build test

all: build

build:
	go build -o ./bin/hotswap hotswap/cmd

test: export GO_UPGRADE_TEST_RESOURCES := $(PWD)/test
test:
	@echo $(VENDOR)
	go test -test.v $(TEST_PKGS)
