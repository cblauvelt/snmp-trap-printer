BINARY     := snmp-trap-printer
BUILD_DIR  := build
MODULE     := github.com/cblauvelt/snmp-trap-printer
CMD        := ./cmd/$(BINARY)/...

.PHONY: help build docker run test test-integration lint clean

## help: Show this help message
help:
	@echo "Usage: make <target>"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## /  /'

## build: Compile the binary to build/
build:
	go build -o $(BUILD_DIR)/$(BINARY) $(CMD)

## docker: Build the Docker image and tag as latest
docker:
	docker build -f deploy/Dockerfile -t $(BINARY):latest .

## run: Build and run with --port 10162 (no sudo required)
run: build
	./$(BUILD_DIR)/$(BINARY) --port 10162

## test: Run all tests
test:
	go test ./...

## test-integration: Run unit and integration tests
test-integration:
	go test -tags integration ./...

## lint: Run go vet
lint:
	go vet ./...

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)
