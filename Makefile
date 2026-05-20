BINARY     := replay
DEV_BINARY := replay-dev
CMD        := ./cmd/replay
GOBIN      ?= $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN      := $(shell go env GOPATH)/bin
endif
VERSION    := $(shell git describe --tags --always --dirty)
LDFLAGS    := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test coverage run clean install

build:
	go build $(LDFLAGS) -o $(BINARY) $(CMD)

test:
	go test ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

run: build
	./$(BINARY) $(ARGS)

install:
	go build $(LDFLAGS) -o $(GOBIN)/$(DEV_BINARY) $(CMD)
	@echo "Installed $(DEV_BINARY) $(VERSION) → $(GOBIN)/$(DEV_BINARY)"

clean:
	rm -f $(BINARY) coverage.out
