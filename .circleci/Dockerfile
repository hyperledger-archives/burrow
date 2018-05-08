FROM golang:1.10.1-alpine3.7
MAINTAINER Monax <support@monax.io>

ENV DOCKER_VERSION "17.12.1-ce"
# This is the image used by the Circle CI config in this directory pushed to quay.io/monax/bosmarmot:ci
# docker build -t quay.io/monax/burrow:ci -f ./.circleci/Dockerfile .
RUN apk add --update --no-cache nodejs netcat-openbsd git openssh-client make bash gcc g++ jq curl
# get docker client
RUN curl -L -o /tmp/docker-$DOCKER_VERSION.tgz https://download.docker.com/linux/static/stable/x86_64/docker-$DOCKER_VERSION.tgz \
    && tar -xz -C /tmp -f /tmp/docker-$DOCKER_VERSION.tgz \
    && mv /tmp/docker/docker /usr/bin \
    && rm -rf /tmp/docker-$DOCKER_VERSION /tmp/docker
RUN go get github.com/jstemmer/go-junit-report
RUN go get -u github.com/golang/dep/cmd/dep
RUN npm install -g mocha
RUN npm install -g mocha-circleci-reporter