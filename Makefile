OS :=

all: monastery

monastery:
	@GOOS=$(OS) go build ./internal/...
	@GOOS=$(OS) go build ./cmd/...

clean:
	@go clean ./...

.PHONY: all monastery clean
