# Function to get the current branch
current_branch = $(shell git rev-parse --abbrev-ref HEAD)

# Function to get the latest tag
latest_tag = $(shell git describe --tags `git rev-list --tags --max-count=1`)

# Function to get the short git SHA
short_sha = $(shell git rev-parse --short HEAD)

# Function to increment the minor version
increment_minor = $(shell echo $1 | awk -F. '{printf "%d.%d.%d", $$1, $$2, $$3+1}')

# Calculate the version
VERSION = \
    $(eval TAG := $(call latest_tag)) \
    $(eval NEW_VERSION := $(call increment_minor,$(subst v,,$(TAG)))) \
    $(if $(findstring main,$(call current_branch)), \
        $(NEW_VERSION), \
        $(NEW_VERSION)-$(call short_sha))

BINARY_NAME=secrets
ORG=OpenPaasDev
REPO=secrets
DATE_TIME := $(shell date '+%Y-%m-%d %H:%M:%S')
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go clean -testcache
	go test ./... -race -covermode=atomic -coverprofile=coverage.out

.PHONY: coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
.PHONY: version
version:
	@echo $(VERSION)

.PHONY: build
build:
	go build -ldflags "-X 'github.com/$(ORG)/$(REPO)/pkg/telemetry.Version=$(VERSION)' -X 'github.com/$(ORG)/$(REPO)/pkg/telemetry.BuildDate=$(DATE_TIME)' -X 'github.com/$(ORG)/$(REPO)/pkg/telemetry.ServiceName=$(BINARY_NAME)'" -o $(BINARY_NAME) main.go

