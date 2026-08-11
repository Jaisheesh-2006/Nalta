GOCMD := go
GOTEST := $(GOCMD) test ./...
GOBUILD := $(GOCMD) build

.PHONY: all build-server test fmt vet run-integration integration-test

all: build-server test

build-server:
	$(GOBUILD) -o nalta .

test:
	$(GOTEST)

fmt:
	$(GOCMD) fmt ./...

vet:
	$(GOCMD) vet ./...

run-integration:
	./integration/run_integration.sh

integration-test:
	$(GOCMD) test ./integration -tags=integration -v
