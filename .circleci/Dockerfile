FROM golang:1.10.3-alpine3.8
MAINTAINER Monax <support@monax.io>

ENV DOCKER_VERSION "17.12.1-ce"
ENV GORELEASER_VERSION "v0.83.0"
# This is the image used by the Circle CI config in this directory pushed to quay.io/monax/bosmarmot:ci
# docker build -t quay.io/monax/build:burrow-ci -f ./.circleci/Dockerfile .
RUN apk add --update --no-cache nodejs npm netcat-openbsd git openssh-client openssl make bash gcc g++ jq curl parallel
RUN echo -ne "will cite" | parallel --citation || true
# get docker client
WORKDIR /usr/bin
RUN curl -L https://download.docker.com/linux/static/stable/x86_64/docker-$DOCKER_VERSION.tgz | tar xz --strip-components 1 docker/docker
RUN curl -L https://github.com/goreleaser/goreleaser/releases/download/$GORELEASER_VERSION/goreleaser_Linux_x86_64.tar.gz | tar xz goreleaser
RUN npm install -g mocha
RUN npm install -g mocha-circleci-reporter
WORKDIR /go/src/github.com/hyperledger/burrow
