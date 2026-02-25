BINARY := replay
CMD    := ./cmd/replay

.PHONY: build test coverage run clean

build:
	go build -o $(BINARY) $(CMD)

test:
	go test ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

run: build
	./$(BINARY) $(ARGS)

clean:
	rm -f $(BINARY) coverage.out
