test:
	go test -v ./...

check: lint vet fmtcheck ineffassign

lint:
	golint -set_exit_status .

vet:
	go vet

fmtcheck:
	@ export output="$$(gofmt -s -d .)"; \
		[ -n "$${output}" ] && echo "$${output}" && export status=1; \
		exit $${status:-0}

ineffassign:
	ineffassign .

setup:
	go get github.com/gordonklaus/ineffassign
	go get github.com/golang/lint/golint
	go get -t -u ./...

.PHONY: test check lint vet fmtcheck ineffassign
