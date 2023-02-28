SHELL := /bin/bash

.PHONY: help
## help: shows this help message
help:
	@ echo "Usage: make [target]"
	@ sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'

.PHONY: run
## run: runs web crawler app. (make run URL=<some_url> DEST_DIR=<some_dir>)
run:
	@ if [ -z "$(URL)" ]; then echo >&2 please set url via the variable URL; exit 2; fi
	@ if [ -z "$(DEST_DIR)" ]; then echo >&2 please set dest dir via the variable DEST_DIR; exit 2; fi
	@ go run cmd/main.go -u $(URL) -d $(DEST_DIR)

.PHONY: vul-setup
## vul-setup: installs Golang's vulnerability check tool
vul-setup:
	@ if [ -z "$$(which govulncheck)" ]; then echo "Installing Golang's vulnerability detection tool..."; go install golang.org/x/vuln/cmd/govulncheck@latest; fi

.PHONY: vul-check
## vul-check: checks for any known vulnerabilities
vul-check: vul-setup
	@ govulncheck ./...

.PHONY: lint
## lint: runs linter
lint: 
	@ docker run  --rm -v "`pwd`:/workspace:cached" -w "/workspace/." golangci/golangci-lint:latest golangci-lint run

.PHONY: test
## test: runs unit tests
test:
	@ go test -cover -v ./... -count=1

.PHONY: coverage
## coverage: run unit tests and generate coverage report in html format
coverage:
	@ go test -coverprofile=coverage.out ./...  && go tool cover -html=coverage.out

.PHONY: int-test
## int-test: runs integration test
int-test:
	@ go test -v ./integration_test --tags=integration
	@ rm -rf integration_test/saved