BINARY     := replay
DEV_BINARY := replay-dev
CMD        := ./cmd/replay
GOBIN      ?= $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN      := $(shell go env GOPATH)/bin
endif

# VERSION: real tag for clean trees, "dev" when there are uncommitted changes.
GIT_DIRTY  := $(shell git status --porcelain)
ifeq ($(GIT_DIRTY),)
VERSION    := $(shell git describe --tags --always)
else
VERSION    := dev
endif

LDFLAGS    := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test coverage run clean install-dev

build:
	go build $(LDFLAGS) -o $(BINARY) $(CMD)

test:
	go test ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

run: build
	./$(BINARY) $(ARGS)

install-dev:
	go build $(LDFLAGS) -o $(GOBIN)/$(DEV_BINARY) $(CMD)
	@echo "Installed $(DEV_BINARY) $(VERSION) → $(GOBIN)/$(DEV_BINARY)"

clean:
	rm -f $(BINARY) coverage.out
