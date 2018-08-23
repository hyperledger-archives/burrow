# ----------------------------------------------------------
# REQUIREMENTS

# - go installed locally

# ----------------------------------------------------------

SHELL := /bin/bash
REPO := $(shell pwd)
GOFILES_NOVENDOR := $(shell go list -f "{{.Dir}}" ./...)
PACKAGES_NOVENDOR := $(shell go list ./...)
# Bosmarmot integration testing
BOSMARMOT_PROJECT := github.com/monax/bosmarmot
BOSMARMOT_GOPATH := ${REPO}/.gopath_bos
BOSMARMOT_CHECKOUT := ${BOSMARMOT_GOPATH}/src/${BOSMARMOT_PROJECT}

# Protobuf generated go files
PROTO_FILES = $(shell find . -path ./vendor -prune -o -path ./.gopath_bos -prune -o -type f -name '*.proto' -print)
PROTO_GO_FILES = $(patsubst %.proto, %.pb.go, $(PROTO_FILES))
PROTO_GO_FILES_REAL = $(shell find . -path ./vendor -prune -o -type f -name '*.pb.go' -print)

# Our own Go files containing the compiled bytecode of solidity files as a constant
SOLIDITY_FILES = $(shell find . -path ./vendor -prune -o -path ./tests -prune -o -type f -name '*.sol' -print)
SOLIDITY_GO_FILES = $(patsubst %.sol, %.sol.go, $(SOLIDITY_FILES))

CI_IMAGE="hyperledger/burrow:ci"

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

# Protobuffing
.PHONY: protobuf_deps
protobuf_deps:
	@go get -u github.com/gogo/protobuf/protoc-gen-gogo
#	@go get -u github.com/golang/protobuf/protoc-gen-go

# Implicit compile rule for GRPC/proto files (note since pb.go files no longer generated
# in same directory as proto file this just regenerates everything
%.pb.go: %.proto
	protoc -I protobuf -I vendor $< --gogo_out=plugins=grpc:${GOPATH}/src

.PHONY: protobuf
protobuf: $(PROTO_GO_FILES)

.PHONY: clean_protobuf
clean_protobuf:
	@rm -f $(PROTO_GO_FILES_REAL)

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
.PHONY: ensure_vendor protobuf
ensure_vendor: reinstall_vendor
	@scripts/is_checkout_dirty.sh

### Building github.com/hyperledger/burrow

# Output commit_hash but only if we have the git repo (e.g. not in docker build
.PHONY: commit_hash
commit_hash:
	@git status &> /dev/null && scripts/commit_hash.sh > commit_hash.txt || true

# build all targets in github.com/hyperledger/burrow
.PHONY: build
build:	check build_burrow

# build all targets in github.com/hyperledger/burrow with checks for race conditions
.PHONY: build_race
build_race:	check build_race_db

# build burrow
.PHONY: build_burrow
build_burrow: commit_hash
	go build -ldflags "-extldflags '-static' \
	-X github.com/hyperledger/burrow/project.commit=$(shell cat commit_hash.txt)" \
	-o ${REPO}/bin/burrow ./cmd/burrow

.PHONY: install_burrow
install_burrow: build_burrow
	cp ${REPO}/bin/burrow ${GOPATH}/bin/burrow

# build burrow with checks for race conditions
.PHONY: build_race_db
build_race_db:
	go build -race -o ${REPO}/bin/burrow ./cmd/burrow

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

# Solidity fixtures
%.sol.go: %.sol scripts/solc_compile_go.sh
	scripts/solc_compile_go.sh $< $@

.PHONY: solidity
solidity: $(SOLIDITY_GO_FILES)

# Test

.PHONY: test
test: check bin/solc
	@tests/scripts/bin_wrapper.sh go test ./... ${GOPACKAGES_NOVENDOR}

.PHONY: test_keys
test_keys: build_burrow
	burrow_bin="${REPO}/bin/burrow" keys/test.sh

rpc/test/strange_loop.go: integration/rpctest
	@rpc/test/strange_loop.sh

# Go will attempt to run separate packages in parallel
.PHONY: test_integration
test_integration: test_keys test_deploy
	@go test -v -tags integration ./integration/...

.PHONY: test_deploy
test_deploy: bin/solc
	@tests/scripts/bin_wrapper.sh tests/deploy.sh

# Run integration test from bosmarmot (separated from other integration tests so we can
# make exception when this test fails when we make a breaking change in Burrow)
.PHONY: test_integration_bosmarmot
test_integration_bosmarmot: bos build_burrow
	cd "${BOSMARMOT_CHECKOUT}" &&\
	make npm_install && \
	GOPATH="${BOSMARMOT_GOPATH}" \
	burrow_bin="${REPO}/bin/burrow" \
	make test_burrow_js_no_burrow

bin/solc: ./tests/scripts/deps/solc.sh
	@mkdir -p bin
	@tests/scripts/deps/solc.sh bin/solc
	@touch bin/solc

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

.PHONY: release
release: docs check test docker_build
	@scripts/is_checkout_dirty.sh || (echo "checkout is dirty so not releasing!" && exit 1)
	@scripts/release.sh

.PHONY: release_dev
release_dev: test docker_build
	@scripts/release_dev.sh

.PHONY: build_ci_image
build_ci_image:
	docker build -t ${CI_IMAGE} -f ./.circleci/Dockerfile .

.PHONY: push_ci_image
push_ci_image: build_ci_image
	docker push ${CI_IMAGE}
