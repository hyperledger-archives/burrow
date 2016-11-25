# TARGET = eris-db
# IMAGE = quay.io/eris/db

SHELL := /bin/bash
GOFILES_NOVENDOR := $(shell find . -type f -name '*.go' -not -path "./vendor/*")
PACKAGES_NOVENDOR := $(shell go list github.com/eris-ltd/eris-db/... | grep -v /vendor/)
VERSION_MIN := $(shell cat ./version/version.go | tail -n 1 | cut -d \  -f 4 | tr -d '"')
VERSION_MAJ := $(shell echo ${VERSION_MIN} | cut -d . -f 1-2)
BUILD_DIR?=target

.PHONY: greet
greet:
	@echo "Hi! I'm the marmot that will help you with eris-db v${VERSION_MIN}"

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

### Building code

# build all targets in github.com/eris-ltd/eris-db
.PHONY: build
build:	build_db build_client build_keys

# build eris-db
.PHONY: build_db
build_db:
	go build -o ${BUILD_DIR}/eris-db ./cmd/eris-db

# build eris-client
.PHONY: build_client
build_client:
	go build -o ${BUILD_DIR}/eris-client ./client/cmd/eris-client

# build eris-keys
.PHONY: build_keys
build_keys:
	@echo "Marmots need to complete moving repository eris-keys into eris-db."
