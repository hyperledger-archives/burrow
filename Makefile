# ----------------------------------------------------------
# REQUIREMENTS

# - go installed locally
# - for build_docker: docker installed locally

# ----------------------------------------------------------

SHELL := /bin/bash
REPO := $(shell pwd)
GOFILES_NOVENDOR := $(shell find ${REPO} -type f -name '*.go' -not -path "${REPO}/vendor/*")
PACKAGES_NOVENDOR := $(shell go list ./... | grep -vF /vendor/)
COMMIT_SHA := $(shell echo `git rev-parse --short --verify HEAD`)

DOCKER_NAMESPACE := quay.io/monax

### Formatting, linting and vetting

# check the code for style standards; currently enforces go formatting.
# display output first, then check for success	
.PHONY: check
check:
	@echo "Checking code for formatting style compliance."
	@gofmt -l -d ${GOFILES_NOVENDOR}
	@gofmt -l ${GOFILES_NOVENDOR} | read && echo && echo "Your marmot has found a problem with the formatting style of the code." 1>&2 && exit 1 || true

# Just fix it
.PHONY: fix
fix:
	@goimports -l -w ${GOFILES_NOVENDOR}

# fmt runs gofmt -w on the code, modifying any files that do not match
# the style guide.
.PHONY: fmt
fmt:
	@echo "Correcting any formatting style corrections."
	@gofmt -l -w ${GOFILES_NOVENDOR}

# lint installs golint and prints recommendations for coding style.
lint: 
	@echo "Running lint checks."
	go get -u github.com/golang/lint/golint
	@for file in $(GOFILES_NOVENDOR); do \
		echo; \
		golint --set_exit_status $${file}; \
	done

# vet runs extended compilation checks to find recommendations for
# suspicious code constructs.
.PHONY: vet
vet:
	@echo "Running go vet."
	@go vet ${PACKAGES_NOVENDOR}

# run the megacheck tool for code compliance
.PHONY: megacheck
megacheck:
	@go get honnef.co/go/tools/cmd/megacheck
	@for pkg in ${PACKAGES_NOVENDOR}; do megacheck "$$pkg"; done

### Dependency management for github.com/hyperledger/burrow

# erase vendor wipes the full vendor directory
.PHONY: erase_vendor
erase_vendor:
	rm -rf ${REPO}/vendor/

# install vendor uses dep to install vendored dependencies
.PHONY: reinstall_vendor
reinstall_vendor: erase_vendor
	@go get -u github.com/golang/dep/cmd/dep
	@dep ensure -v

# delete the vendor directy and pull back using dep lock and constraints file
# will exit with an error if the working directory is not clean (any missing files or new
# untracked ones)
.PHONY: ensure_vendor
ensure_vendor: reinstall_vendor
	@scripts/is_checkout_dirty.sh

# dumps Solidity interface contracts for SNatives
.PHONY: snatives
snatives:
	@go run ./util/snatives/cmd/main.go

### Building github.com/hyperledger/burrow

# build all targets in github.com/hyperledger/burrow
.PHONY: build
build:	check build_db build_client

# build all targets in github.com/hyperledger/burrow with checks for race conditions
.PHONY: build_race
build_race:	check build_race_db build_race_client build_race_keys

# build burrow
.PHONY: build_db
build_db:
	go build -o ${REPO}/target/burrow-${COMMIT_SHA} ./cmd/burrow

# build burrow-client
.PHONY: build_client
build_client:
	go build -o ${REPO}/target/burrow-client-${COMMIT_SHA} ./client/cmd/burrow-client

# build burrow with checks for race conditions
.PHONY: build_race_db
build_race_db:
	go build -race -o ${REPO}/target/burrow-${COMMIT_SHA} ./cmd/burrow

# build burrow-client with checks for race conditions
.PHONY: build_race_client
build_race_client:
	go build -race -o ${REPO}/target/burrow-client-${COMMIT_SHA} ./client/cmd/burrow-client

### Build docker images for github.com/hyperledger/burrow

# build docker image for burrow
.PHONY: build_docker_db
build_docker_db: check
	@scripts/build_tool.sh

### Testing github.com/hyperledger/burrow

# test burrow
.PHONY: test
test: check
	@go test ${PACKAGES_NOVENDOR}

.PHONY: test_integration
test_integration:
	@go get -u github.com/monax/keys/cmd/monax-keys
	@go test ./keys/integration -tags integration
	@go test ./rpc/tm/integration -tags integration

# test burrow with checks for race conditions
.PHONY: test_race
test_race: build_race
	@go test -race ${PACKAGES_NOVENDOR}

### Clean up

# clean removes the target folder containing build artefacts
.PHONY: clean
clean:
	-rm -r ./target

### Release and versioning

# Print version
.PHONY: version
version:
	@go run ./project/cmd/version/main.go

# Generate full changelog of all release notes
CHANGELOG.md: project/history.go project/cmd/changelog/main.go
	@go run ./project/cmd/changelog/main.go > CHANGELOG.md

# Generated release note for this version
NOTES.md: project/history.go project/cmd/notes/main.go
	@go run ./project/cmd/notes/main.go > NOTES.md

.PHONY: docs
docs: CHANGELOG.md NOTES.md

# Tag the current HEAD commit with the current release defined in
# ./release/release.go
.PHONY: tag_release
tag_release: test check CHANGELOG.md NOTES.md build
	@scripts/tag_release.sh

