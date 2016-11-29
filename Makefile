# ----------------------------------------------------------
# REQUIREMENTS

# - go installed locally
# - for build_docker: docker installed locally

# ----------------------------------------------------------

SHELL := /bin/bash
REPO := $(shell pwd)
GOFILES_NOVENDOR := $(shell find ${REPO} -type f -name '*.go' -not -path "${REPO}/vendor/*")
PACKAGES_NOVENDOR := $(shell go list github.com/eris-ltd/eris-db/... | grep -v /vendor/)
VERSION := $(shell cat ${REPO}/version/version.go | tail -n 1 | cut -d \  -f 4 | tr -d '"')
VERSION_MIN := $(shell echo ${VERSION} | cut -d . -f 1-2)
COMMIT_SHA := $(shell echo `git rev-parse --short --verify HEAD`)

DOCKER_NAMESPACE := quay.io/eris


.PHONY: greet
greet:
	@echo "Hi! I'm the marmot that will help you with eris-db v${VERSION}"

### Formatting, linting and vetting

# check the code for style standards; currently enforces go formatting.
# display output first, then check for success	
.PHONY: check
check:
	@echo "Checking code for formatting style compliance."
	@gofmt -l -d ${GOFILES_NOVENDOR}
	@gofmt -l ${GOFILES_NOVENDOR} | read && echo && echo "Your marmot has found a problem with the formatting style of the code." 1>&2 && exit 1 || true

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

### Dependency management for github.com/eris-ltd/eris-db

# erase vendor wipes the full vendor directory
.PHONY: erase_vendor
erase_vendor:
	rm -rf ${REPO}/vendor/

# install vendor uses glide to install vendored dependencies
.PHONY: install_vendor
install_vendor:
	go get github.com/Masterminds/glide
	glide install

### Building github.com/eris-ltd/eris-db

# build all targets in github.com/eris-ltd/eris-db
.PHONY: build
build:	check build_db build_client build_keys

# build all targets in github.com/eris-ltd/eris-db with checks for race conditions
.PHONY: build_race
build_race:	check build_race_db build_race_client build_race_keys

# build eris-db
.PHONY: build_db
build_db:
	go build -o ${REPO}/target/eris-db-${COMMIT_SHA} ./cmd/eris-db

# build eris-client
.PHONY: build_client
build_client:
	go build -o ${REPO}/target/eris-client-${COMMIT_SHA} ./client/cmd/eris-client

# build eris-keys
.PHONY: build_keys
build_keys:
	@echo "Marmots need to complete moving repository eris-keys into eris-db."

# build eris-db with checks for race conditions
.PHONY: build_race_db
build_race_db:
	go build -race -o ${REPO}/target/eris-db-${COMMIT_SHA} ./cmd/eris-db

# build eris-client with checks for race conditions
.PHONY: build_race_client
build_race_client:
	go build -race -o ${REPO}/target/eris-client-${COMMIT_SHA} ./client/cmd/eris-client

# build eris-keys with checks for race conditions
.PHONY: build_race_keys
build_race_keys:
	@echo "Marmots need to complete moving repository eris-keys into eris-db."

### Testing github.com/eris-ltd/eris-db

# test eris-db
.PHONY: test
test: build
	@go test ${PACKAGES_NOVENDOR}

# test eris-db with checks for race conditions
.PHONY: test_race
test_race: build_race
	@go test -race ${PACKAGES_NOVENDOR}

### Build docker images for github.com/eris-ltd/eris-db

# build docker image for eris-db
.PHONY: build_docker_db
build_docker_db: check
	@mkdir -p ${REPO}/target/docker
	docker build -t ${DOCKER_NAMESPACE}/db:build-${COMMIT_SHA} ${REPO}
	docker run --rm --entrypoint cat ${DOCKER_NAMESPACE}/db:build-${COMMIT_SHA} /usr/local/bin/eris-db > ${REPO}/target/docker/eris-db.dockerartefact
	docker run --rm --entrypoint cat ${DOCKER_NAMESPACE}/db:build-${COMMIT_SHA} /usr/local/bin/eris-client > ${REPO}/target/docker/eris-client.dockerartefact
	docker build -t ${DOCKER_NAMESPACE}/db:${VERSION} -f Dockerfile.deploy ${REPO}

	@rm ${REPO}/target/docker/eris-db.dockerartefact
	@rm ${REPO}/target/docker/eris-client.dockerartefact
	docker rmi ${DOCKER_NAMESPACE}/db:build-${COMMIT_SHA}

### Test docker images for github.com/eris-ltd/eris-db

# test docker image for eris-db
.PHONY: test_docker_db
test_docker_db: check
	docker build -t ${DOCKER_NAMESPACE}/db:build-${COMMIT_SHA} ${REPO}
	docker run ${DOCKER_NAMESPACE}/db:build-${COMMIT_SHA} glide nv | xargs go test