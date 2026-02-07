.PHONY: unit unit-race unit-cover test-docker
.PHONY: fmt fmt-check lint lint-docker
.PHONY: mod-tidy mod-check check
.PHONY: build build-release install clean
.PHONY: docker-build docker-run help

.DEFAULT_GOAL := build

PKG := github.com/lightninglabs/lightning-agent-kit/mcp-server

GOTEST := GO111MODULE=on go test -v

GO_BIN := ${GOPATH}/bin
GOBUILD := GO111MODULE=on go build -v
GOINSTALL := GO111MODULE=on go install -v
GOMOD := GO111MODULE=on go mod

COMMIT = $(shell git describe --abbrev=40 --always --dirty)
LDFLAGS := -ldflags "-X $(PKG).Commit=$(COMMIT)"
DEV_TAGS = dev

GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./external/*")
GOLIST := go list $(PKG)/... | grep -v '/vendor/' | grep -v '/external/'

TEST_FLAGS = -test.timeout=20m

UNIT := $(GOLIST) | xargs -L 1 env $(GOTEST) $(TEST_FLAGS)

GREEN := "\\033[0;32m"
NC := "\\033[0m"
define print
	echo $(GREEN)$1$(NC)
endef

# =======
# TESTING
# =======

unit: #? Run unit tests
	@$(call print, "Running unit tests.")
	$(UNIT)

unit-race: #? Run unit tests with race detector
	@$(call print, "Running unit race tests.")
	env CGO_ENABLED=1 GORACE="history_size=7 halt_on_errors=1" $(UNIT) -race

test-docker: #? Run unit tests inside golang container
	@$(call print, "Running unit tests inside golang:1.24.5 container.")
	docker run --rm \
		-v $(PWD):/workspace \
		-w /workspace \
		-e GO111MODULE=on \
		golang:1.24.5 go test ./...

unit-cover: #? Run unit tests with coverage
	@$(call print, "Running unit tests with coverage.")
	$(GOTEST) -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# ==========
# FORMATTING
# ==========

fmt: #? Format code and fix imports
	@$(call print, "Formatting source.")
	gofmt -l -w -s $(GOFILES_NOVENDOR)

fmt-check: #? Verify code is formatted correctly
	@$(call print, "Checking formatting.")
	@gofmt -l $(GOFILES_NOVENDOR) | tee /tmp/fmt-issues
	@test ! -s /tmp/fmt-issues || (echo "Code is not formatted. Please run 'make fmt'" && exit 1)

lint: #? Run golangci-lint
	@$(call print, "Linting source.")
	golangci-lint run -v

lint-docker: #? Run golangci-lint in Docker container
	@$(call print, "Running linter in Docker container.")
	docker build -f Dockerfile.lint -t mcp-lnc-lint .
	docker run --rm mcp-lnc-lint

# =======
# MODULES
# =======

mod-tidy: #? Tidy Go modules
	@$(call print, "Tidying modules.")
	$(GOMOD) tidy

mod-check: #? Verify modules are tidy
	@$(call print, "Checking modules.")
	$(GOMOD) tidy
	if test -n "$$(git status --porcelain | grep -e "go.mod\|go.sum")"; then echo "Running go mod tidy changes go.mod/go.sum"; git status; git diff; exit 1; fi

# ============
# BUILDING
# ============

build: #? Build debug binary
	@$(call print, "Building MCP LNC server.")
	$(GOBUILD) -tags="$(DEV_TAGS)" -o mcp-lnc-server $(LDFLAGS) .

build-release: #? Build optimized release binary
	@$(call print, "Building release MCP LNC server.")
	$(GOBUILD) -o mcp-lnc-server $(LDFLAGS) .

install: #? Install binary to GOPATH/bin
	@$(call print, "Installing MCP LNC server.")
	$(GOINSTALL) -tags="${tags}" $(LDFLAGS) .

# =======
# CLEANUP
# =======

clean: #? Remove build artifacts
	@$(call print, "Cleaning build artifacts.")
	rm -f mcp-lnc-server
	rm -f coverage.out coverage.html

# =======
# QUALITY
# =======

check: fmt lint mod-check unit #? Run all checks
	@$(call print, "All checks passed!")

# ==============
# CONTAINERIZATION
# ==============

docker-build: #? Build Docker image
	@$(call print, "Building Docker image.")
	docker build -t mcp-lnc-server .

docker-run: #? Run Docker container
	@$(call print, "Running Docker container.")
	docker run --rm -p 8080:8080 mcp-lnc-server

# =======
# HELP
# =======

help: #? Show available targets
	@echo "Available make targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?#\? .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?#\\? "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make unit"
	@echo "  make unit-race"
	@echo "  make check"
