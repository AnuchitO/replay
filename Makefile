BINARY     := replay
DEV_BINARY := replay-dev
CMD        := ./cmd/replay
GOBIN      ?= $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN      := $(shell go env GOPATH)/bin
endif

.PHONY: build test coverage run clean install

build:
	go build -o $(BINARY) $(CMD)

test:
	go test ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

run: build
	./$(BINARY) $(ARGS)

install:
	go build -o $(GOBIN)/$(DEV_BINARY) $(CMD)
	@echo "Installed $(DEV_BINARY) → $(GOBIN)/$(DEV_BINARY)"

clean:
	rm -f $(BINARY) coverage.out
