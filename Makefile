TARGET = pxl
BINARY = ./bin/$(TARGET)

GO = go

UPX := $(shell upx --version 2>/dev/null)

all: build

build: clean
	@mkdir -p ./bin
	$(GO) build -ldflags="-s -w" -o $(BINARY) ./cli/main.go
ifdef UPX
	upx --brute $(BINARY)
endif

clean:
	rm -f $(BINARY)

test:
	$(GO) test -race ./...

lint:
	golangci-lint run ./...

.PHONY: all build clean test lint
