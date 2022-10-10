GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
BINARY_NAME=me-snort-parser
VERSION?=1.1

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

build: vendor build-linux ## Build the project and put the output binary in out/bin/

build-linux:
	@echo "Compiling for i386"
	@ GOOS=linux GOARCH=386 GO111MODULE=on $(GOCMD) build -mod vendor -o out/bin/$(BINARY_NAME)-linux-i386 .
	@echo "Compiling for amd64"
	@ GOOS=linux GOARCH=amd64 GO111MODULE=on $(GOCMD) build -mod vendor -o out/bin/$(BINARY_NAME)-linux-amd64 .
	@echo "Compiling for arm (32-bit)"
	@ GOOS=linux GOARCH=arm GO111MODULE=on $(GOCMD) build -mod vendor -o out/bin/$(BINARY_NAME)-linux-arm .
	@echo "Compiling for arm64"
	@ GOOS=linux GOARCH=arm64 GO111MODULE=on $(GOCMD) build -mod vendor -o out/bin/$(BINARY_NAME)-linux-arm64 .

clean: ## Remove build related file
	@rm -rf ./bin
	@rm -rf ./out
	@echo "Any build output removed."

vendor: ## Copy of all packages needed to support builds in the vendor directory
	@ $(GOCMD) mod vendor

run: ## Run with go run
	@go run main.go

help: ## Show this help.
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)
