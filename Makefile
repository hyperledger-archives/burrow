# ----------------------------------------------------------
# REQUIREMENTS

# - Go 1.12
# - Make
# - jq
# - find
# - bash
# - protoc (for rebuilding protobuf files)

# ----------------------------------------------------------

SHELL := /bin/bash
REPO := $(shell pwd)
GOFILES := $(shell go list -f "{{.Dir}}" ./...)
PACKAGES := $(shell go list ./... )

# Protobuf generated go files
PROTO_FILES = $(shell find . -path ./vendor -prune -o -path ./.gopath_bos -prune -o -type f -name '*.proto' -print)
PROTO_GO_FILES = $(patsubst %.proto, %.pb.go, $(PROTO_FILES))
PROTO_GO_FILES_REAL = $(shell find . -path ./vendor -prune -o -type f -name '*.pb.go' -print)

# Our own Go files containing the compiled bytecode of solidity files as a constant
SOLIDITY_FILES = $(shell find . -path ./vendor -prune -o -path ./tests -prune -o -type f -name '*.sol' -print)
SOLIDITY_GO_FILES = $(patsubst %.sol, %.sol.go, $(SOLIDITY_FILES))
SOLANG_FILES = $(shell find . -path ./vendor -prune -o -path ./tests -prune -o -type f -name '*.solang' -print)
SOLANG_GO_FILES = $(patsubst %.solang, %.solang.go, $(SOLANG_FILES))

CI_IMAGE="hyperledger/burrow:ci"

GOPATH?=${HOME}/go
BIN_PATH?=${GOPATH}/bin

export GO111MODULE=on

### Formatting, linting and vetting

# check the code for style standards; currently enforces go formatting.
# display output first, then check for success
.PHONY: check
check:
	@echo "Checking code for formatting style compliance."
	@gofmt -l -d ${GOFILES}
	@gofmt -l ${GOFILES} | read && echo && echo "Your marmot has found a problem with the formatting style of the code." 1>&2 && exit 1 || true

# Just fix it
.PHONY: fix
fix:
	@goimports -l -w ${GOFILES}

# fmt runs gofmt -w on the code, modifying any files that do not match
# the style guide.
.PHONY: fmt
fmt:
	@echo "Correcting any formatting style corrections."
	@gofmt -l -w ${GOFILES}

# lint installs golint and prints recommendations for coding style.
lint:
	@echo "Running lint checks."
	go get -u github.com/golang/lint/golint
	@for file in $(GOFILES); do \
		echo; \
		golint --set_exit_status $${file}; \
	done

# vet runs extended compilation checks to find recommendations for
# suspicious code constructs.
.PHONY: vet
vet:
	@echo "Running go vet."
	@go vet ${PACKAGES}

# run the megacheck tool for code compliance
.PHONY: megacheck
megacheck:
	@go get honnef.co/go/tools/cmd/megacheck
	@for pkg in ${PACKAGES}; do megacheck "$$pkg"; done

# Protobuffing
.PHONY: protobuf_deps
protobuf_deps:
	@go get -u github.com/gogo/protobuf/protoc-gen-gogo
#	@go get -u github.com/golang/protobuf/protoc-gen-go

# Implicit compile rule for GRPC/proto files (note since pb.go files no longer generated
# in same directory as proto file this just regenerates everything
%.pb.go: %.proto
	protoc -I ./protobuf $< --gogo_out=plugins=grpc:${GOPATH}/src

.PHONY: protobuf
protobuf: $(PROTO_GO_FILES)

.PHONY: clean_protobuf
clean_protobuf:
	@rm -f $(PROTO_GO_FILES_REAL)


### PEG query grammar

# This allows us to filter tagged objects with things like (EventID = 'foo' OR Height > 10) AND EventName CONTAINS 'frog'

.PHONY: peg_deps
peg_deps:
	go get -u github.com/pointlander/peg

# regenerate the parser
.PHONY: peg
peg:
	peg event/query/query.peg

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

# build burrow and vent
.PHONY: build_burrow
build_burrow: commit_hash
	go build -ldflags "-extldflags '-static' \
	-X github.com/hyperledger/burrow/project.commit=$(shell cat commit_hash.txt) \
	-X github.com/hyperledger/burrow/project.date=$(shell date '+%Y-%m-%d')" \
	-o ${REPO}/bin/burrow ./cmd/burrow

# With the sqlite tag - enabling Vent sqlite adapter support, but building a CGO binary
.PHONY: build_burrow_sqlite
build_burrow_sqlite: commit_hash
	go build -tags sqlite \
	 -ldflags "-extldflags '-static' \
	-X github.com/hyperledger/burrow/project.commit=$(shell cat commit_hash.txt) \
	-X github.com/hyperledger/burrow/project.date=$(shell date -I)" \
	-o ${REPO}/bin/burrow-vent-sqlite ./cmd/burrow

.PHONY: install
install: build_burrow
	mkdir -p ${BIN_PATH}
	install -T ${REPO}/bin/burrow ${BIN_PATH}/burrow

# build burrow with checks for race conditions
.PHONY: build_race_db
build_race_db:
	go build -race -o ${REPO}/bin/burrow ./cmd/burrow

### Build docker images for github.com/hyperledger/burrow

# build docker image for burrow
.PHONY: docker_build
docker_build: check commit_hash
	@scripts/build_tool.sh

### Testing github.com/hyperledger/burrow

# Solidity fixtures
%.sol.go: %.sol
	@go run ./deploy/compile/solgo/main.go $^

# Solidity fixtures
%.solang.go: %.solang
	@go run ./deploy/compile/solgo/main.go -wasm $^

.PHONY: solidity
solidity: $(SOLIDITY_GO_FILES)

.PHONY: solang
solang: $(SOLANG_GO_FILES)

# node/js
#
# Install dependency
.PHONY: npm_install
npm_install:
	npm install

.PHONY: test_js
test_js: bin/solc build_burrow
	./tests/scripts/bin_wrapper.sh npm test

# Test

.PHONY: test
test: check bin/solc
	@tests/scripts/bin_wrapper.sh go test ./... ${GO_TEST_ARGS}

.PHONY: test_keys
test_keys: build_burrow
	burrow_bin="${REPO}/bin/burrow" tests/keys_server/test.sh

.PHONY:	test_truffle
test_truffle: build_burrow
	burrow_bin="${REPO}/bin/burrow" tests/web3/truffle.sh

.PHONY:	test_integration_vent
test_integration_vent:
	# Include sqlite adapter with tests - will build with CGO but that's probably fine
	go test -v -tags 'integration sqlite' ./vent/...

.PHONY:	test_integration_vent_postgres
test_integration_vent_postgres:
	docker-compose run burrow make test_integration_vent

.PHONY: test_restore
test_restore: build_burrow bin/solc
	@tests/scripts/bin_wrapper.sh tests/dump/test.sh

# Go will attempt to run separate packages in parallel
.PHONY: test_integration
test_integration: test_keys test_deploy test_integration_vent_postgres test_restore test_truffle
	@go test -v -tags integration ./integration/...

.PHONY: test_integration_no_postgres
test_integration_no_postgres: test_keys test_deploy test_integration_vent test_restore test_truffle
	@go test -v -tags integration ./integration/...

.PHONY: test_deploy
test_deploy: bin/solc build_burrow
	@tests/scripts/bin_wrapper.sh tests/deploy.sh

bin/solc: ./tests/scripts/deps/solc.sh
	@mkdir -p bin
	@tests/scripts/deps/solc.sh bin/solc
	@touch bin/solc

# test burrow with checks for race conditions
.PHONY: test_race
test_race: build_race
	@go test -race ${PACKAGES}

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
# ./project/history.go
.PHONY: tag_release
tag_release: test check docs build
	@scripts/tag_release.sh

.PHONY: build_ci_image
build_ci_image:
	docker build -t ${CI_IMAGE} -f ./.circleci/Dockerfile .

.PHONY: push_ci_image
push_ci_image: build_ci_image
	docker push ${CI_IMAGE}

.PHONY: ready_for_pull_request
ready_for_pull_request: docs fix

.PHONY: staticcheck
staticcheck:
	go get honnef.co/go/tools/cmd/staticcheck
	staticcheck ./...
