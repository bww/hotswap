
export VENDOR := $(shell cd ../../../.. && pwd)
export GOPATH := $(GOPATH):$(VENDOR)

TEST_PKGS := github.com/bww/go-upgrade github.com/bww/go-upgrade/driver/postgres

.PHONY: all test

all: test

test: export GO_UPGRADE_TEST_RESOURCES := $(PWD)/test
test:
	@echo $(VENDOR)
	go test -test.v $(TEST_PKGS)
