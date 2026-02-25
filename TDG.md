# TDG Configuration

## Project Information
- Language: Go
- Framework: None (CLI tool)
- Test Framework: Go standard testing package

## Build Command
```bash
go build ./...
```

## Test Command
```bash
go test ./...
```

## Single Test Command
```bash
go test -run ^TestName$ ./path/to/package
```

## Coverage Command
```bash
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
```

## Test File Patterns
- Test files: `*_test.go`
- Test directory: Co-located with source files
