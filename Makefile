BIN     := hncli
CMD     := ./cmd/hncli
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: all build install clean test vet fmt

all: build

build:
	go build $(LDFLAGS) -o $(BIN) $(CMD)

install:
	go install $(LDFLAGS) $(CMD)

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

clean:
	rm -f $(BIN)
