all: monastery

monastery:
	@go build ./internal/...
	@go build ./cmd/...

clean:
	@go clean ./...

.PHONY: all monastery clean
