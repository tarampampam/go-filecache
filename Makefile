#!/usr/bin/make
# Makefile readme (ru): <http://linux.yaroslavl.ru/docs/prog/gnu_make_3-79_russian_manual.html>
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh
LDFLAGS = "-s -w"

DOCKER_BIN = $(shell command -v docker 2> /dev/null)
DC_BIN = $(shell command -v docker-compose 2> /dev/null)
DC_RUN_ARGS = --rm --user "$(shell id -u):$(shell id -g)" app

.PHONY : help fmt lint gotest gobench test cover 'shell' clean
.DEFAULT_GOAL : help
.SILENT : test 'shell'

# This will output the help for each task. thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Show this help
	@printf "\033[33m%s:\033[0m\n" 'Available commands'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[32m%-11s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

fmt: ## Run source code formatter tools
	$(DC_BIN) run $(DC_RUN_ARGS) sh -c 'GO111MODULE=off go get golang.org/x/tools/cmd/goimports && $$GOPATH/bin/goimports -d -w .'
	$(DC_BIN) run $(DC_RUN_ARGS) gofmt -s -w -d .

lint: ## Run package linters
	$(DOCKER_BIN) run --rm -t -v $(shell pwd):/app -w /app golangci/golangci-lint:latest-alpine golangci-lint run -v

gotest: ## Run package tests
	$(DC_BIN) run $(DC_RUN_ARGS) go test -v -race -timeout 5s ./...

gobench: ## Run package benchmarks
	$(DC_BIN) run $(DC_RUN_ARGS) go test -benchmem -bench=.

test: lint gotest ## Run package tests and linters
	@printf "\n   \e[30;42m %s \033[0m\n\n" 'All tests passed!';

cover: ## Run package tests with coverage report
	$(DC_BIN) run $(DC_RUN_ARGS) sh -c 'go test -race -covermode=atomic -coverprofile /tmp/cp.out ./... && go tool cover -html=/tmp/cp.out -o ./coverage.html'
	-sensible-browser ./coverage.html && sleep 2 && rm -f ./coverage.html

shell: ## Start shell into container with golang
	$(DC_BIN) run $(DC_RUN_ARGS) bash

clean: ## Make clean
	$(DC_BIN) down -v -t 1
