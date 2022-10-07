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

build: build-linux ## Build the project and put the output binary in out/bin/
	mkdir -p out/bin
	echo "Compiling snort-parser"

build-linux:
	for arch in {386,amd64,arm64}; do\
		GOOS=linux GOARCH=$${arch} GO111MODULE=on $(GOCMD) build -mod vendor -o out/bin/$(BINARY_NAME)-linux-$${arch} . ;\
	done

clean: ## Remove build related file
	rm -rf ./bin
	rm -rf ./out

vendor: ## Copy of all packages needed to support builds in the vendor directory
	$(GOCMD) mod vendor

run: ## Run with go run
	go run main.go

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
