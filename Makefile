# ----------------------------------------------------------
# REQUIREMENTS

# - go installed locally
# - for build_docker: docker installed locally

# ----------------------------------------------------------

SHELL := /bin/bash
REPO := $(shell pwd)
GOFILES_NOVENDOR := $(shell go list -f "{{.Dir}}" ./...)
PACKAGES_NOVENDOR := $(shell go list ./...)
LDFLAGS :=
# Bosmarmot integration testing
BOSMARMOT_PROJECT := github.com/monax/bosmarmot
BOSMARMOT_GOPATH := ${REPO}/.gopath_bos
BOSMARMOT_CHECKOUT := ${BOSMARMOT_GOPATH}/src/${BOSMARMOT_PROJECT}

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

# Output commit_hash but only if we have the git repo (e.g. not in docker build
.PHONY: commit_hash
commit_hash:
	@git status &> /dev/null && scripts/commit_hash.sh > commit_hash.txt || true

# build all targets in github.com/hyperledger/burrow
.PHONY: build
build:	check build_db build_client

# install burrow
.PHONY: install
install: check build_db build_client 
	cp ./bin/burrow ${GOPATH}/bin
	cp ./bin/burrow-client ${GOPATH}/bin


# build all targets in github.com/hyperledger/burrow with checks for race conditions
.PHONY: build_race
build_race:	check build_race_db build_race_client

# build burrow
.PHONY: build_db
build_db: commit_hash
	go build -ldflags "-extldflags '-static' \
	-X github.com/hyperledger/burrow/project.commit=$(shell cat commit_hash.txt)" \
	-o ${REPO}/bin/burrow ./cmd/burrow

.PHONY: install_db
install_db: build_db
	cp ${REPO}/bin/burrow ${GOPATH}/bin/burrow

# build burrow-client
.PHONY: build_client
build_client: commit_hash
	go build -ldflags "-extldflags '-static' \
	-X github.com/hyperledger/burrow/project.commit=$(shell cat commit_hash.txt)" \
	-o ${REPO}/bin/burrow-client ./client/cmd/burrow-client

# build burrow with checks for race conditions
.PHONY: build_race_db
build_race_db:
	go build -race -o ${REPO}/bin/burrow ./cmd/burrow

# build burrow-client with checks for race conditions
.PHONY: build_race_client
build_race_client:
	go build -race -o ${REPO}/bin/burrow-client ./client/cmd/burrow-client


# Get the Bosmarmot code
.PHONY: bos
bos: ./scripts/deps/bos.sh
	scripts/git_get_revision.sh \
	https://${BOSMARMOT_PROJECT}.git \
	${BOSMARMOT_CHECKOUT} \
	$(shell ./scripts/deps/bos.sh)

### Build docker images for github.com/hyperledger/burrow

# build docker image for burrow
.PHONY: docker_build
docker_build: check commit_hash
	@scripts/build_tool.sh

### Testing github.com/hyperledger/burrow

# test burrow
.PHONY: test
test: check
	@go test ${PACKAGES_NOVENDOR}

.PHONY: test_integration
test_integration:
	@go get github.com/monax/bosmarmot/keys/cmd/monax-keys
	@go test -tags integration ./keys/integration
	@go test -tags integration ./rpc/v0/integration
	@go test -tags integration ./rpc/tm/integration

# Run integration test from bosmarmot (separated from other integration tests so we can
# make exception when this test fails when we make a breaking change in Burrow)
.PHONY: test_integration_bosmarmot
test_integration_bosmarmot: bos build_db
	cd "${BOSMARMOT_CHECKOUT}" &&\
	make npm_install && \
	GOPATH="${BOSMARMOT_GOPATH}" \
	burrow_bin="${REPO}/bin/burrow" \
	make test_integration_no_burrow


# test burrow with checks for race conditions
.PHONY: test_race
test_race: build_race
	@go test -race ${PACKAGES_NOVENDOR}

### Clean up

# clean removes the target folder containing build artefacts
.PHONY: clean
clean:
	-rm -r ./bin

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

