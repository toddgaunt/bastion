all: monastery

#ARCHFLAGS := GOOS=linux GOARCH=arm GOARM=5
ARCHFLAGS :=

monastery:
	$(ARCHFLAGS) go build ./internal/...
	$(ARCHFLAGS) go build ./cmd/...

clean:
	go clean ./...

.PHONY: all monastery clean
