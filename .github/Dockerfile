FROM golang:1.15-alpine3.13
MAINTAINER Monax <support@monax.io>

ENV DOCKER_VERSION "17.12.1-ce"
ENV GORELEASER_VERSION "v0.104.1"
# This is the image used by the Circle CI config
# Update remote with 'make push_ci_image'
RUN apk add --update --no-cache \
 nodejs \
 yarn \
 netcat-openbsd \
 git \
 openssh-client \
 openssl \
 make \
 bash \
 gcc \
 g++ \
 jq \
 curl \
 docker \
 libffi-dev \
 openssl-dev \
 python3-dev \
 py-pip
RUN pip3 install docker-compose
# get docker client
WORKDIR /usr/bin
RUN curl -sS -L https://download.docker.com/linux/static/stable/x86_64/docker-$DOCKER_VERSION.tgz | tar xz --strip-components 1 docker/docker
RUN curl -sS -L https://github.com/goreleaser/goreleaser/releases/download/$GORELEASER_VERSION/goreleaser_Linux_x86_64.tar.gz | tar xz goreleaser
ENV CI=true
# Protobuf builds require being in GOPATH (more or less) but this disables module support by default
ENV GO111MODULE=on
WORKDIR /go/src/github.com/hyperledger/burrow
