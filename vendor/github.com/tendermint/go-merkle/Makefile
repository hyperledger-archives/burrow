.PHONY: get_deps all bench test

all: test

test:
	go test -v -race `glide novendor`

bench:
	go test `glide novendor` -tags gcc -bench=.

get_deps:
	go get github.com/Masterminds/glide
	rm -rf ./vendor
	glide install
