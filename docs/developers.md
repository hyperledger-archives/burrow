# Developers Guide

## Prerequisites

- [Go](https://golang.org/doc/install) (Version >= 1.11)
- [golint](https://github.com/golang/lint)
- [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports)
- [protoc](http://google.github.io/proto-lens/installing-protoc.html) (libprotoc 3.7.1)

Please also refer to our [contributing guidelines](https://github.com/hyperledger/burrow/blob/develop/.github/CONTRIBUTING.md).

## Building

Statically build the burrow binary with `make build` (output in `./bin`) or install to `${GOPATH}/bin` with `make install`.

## Testing

Before submitting a PR, after making any changes, run `make test` to ensure that the unit tests pass and `make test_integration` 
for integration tests. If there are any formatting problems, try to run `make fmt` or `make fix`.

## gRPC and Protobuf

Install protoc and run `make protobuf_deps`. If you make any changes to the protobuf specs, run `make protobuf` to re-compile.

## Releasing

* First of all make sure everyone is happy with doing a release now. 
* Update project/history.go with the latest releases notes and version. Run `make CHANGELOG.md NOTES.md` and make sure this is merged to develop.
* On the develop branch, run `make ready_for_pull_request`. Check for any modified files.
* Using the github.com web interface, create a pull request for master <= develop (so merging latest develop into master)
* Get someone to merge it. They should check that all commits from develop are included using `git log --oneline origin/develop ^origin/master`
* Once master is update to date, switch to master locally run `make tag_release`. This will push the tag which kicks of the release build.
* Optionally send out email on hyperledger burrow mailinglist. Agreements network email should be sent out automatically.

## Proposals

### Architecture Decision Records (ADRs)

ADRs describe standards for the Hyperledger Burrow platform, including core protocol specifications, and client APIs.

### Contributing

 1. Review [ADR-1](ADRs/adr-1.md).
 2. Fork the repository by clicking "Fork" in the top right.
 3. Add your ADR to your fork of the repository. There is a [template ADR here](ADRs/adr-X_template.md).
 4. Submit a Pull Request to Burrow's [ADRs repository](./ADRs/).

If your ADR requires images, the image files should be included in a subdirectory of the `assets` folder for that ADR as follow: `assets/ADR-X` (for ADR **X**). When linking to an image in the ADR, use relative links such as `../assets/adr-X/image.png`.