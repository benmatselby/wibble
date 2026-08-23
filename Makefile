NAME := wibble
BUILD_DIR ?= build
GOOS ?=
ARCH ?=
OUT_PATH=$(BUILD_DIR)/$(NAME)-$(GOOS)-$(GOARCH)
GIT_ORG ?= benmatselby
GIT_RELEASE ?= $(shell git rev-parse --short HEAD)

.PHONY: explain
explain:
	### Welcome
	#
	# ▖ ▗▖▄ ▗▖   ▗▖   █ ▗▞▀▚▖
	# ▌ ▐▌▄ ▐▌   ▐▌   █ ▐▛▀▀▘
	# ▌ ▐▌█ ▐▛▀▚▖▐▛▀▚▖█ ▝▚▄▄▖
	# ▙█▟▌█ ▐▙▄▞▘▐▙▄▞▘█
	#
	### Installation
	#
	# $$ make all
	#
	### Targets
	@cat Makefile* | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: clean
clean: ## Clean the local dependencies
	rm -fr $(BUILD_DIR) && mkdir -p $(BUILD_DIR)

.PHONY: install
install: ## Install the local dependencies
	go get ./...
	go install go.uber.org/mock/mockgen@latest

.PHONY: lint
lint: ## Vet the code
	golangci-lint run

.PHONY: build
build: ## Build the application
	go build -o $(NAME) .

.PHONY:
mocks: ## Generate the mocks
	go generate -x ./...

.PHONY: test
test: mocks ## Run the unit tests
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

.PHONY: test-cov
test-cov: mocks test ## Run the unit tests with coverage
	go tool cover -html=coverage.out -o build/coverage.html

.PHONY: all
all: clean install build test ## Run everything

.PHONY: all-dev
all-dev: clean build lint test ## Run everything (dev)

