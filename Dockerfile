# We use a multistage build to avoid bloating our deployment image with build dependencies
FROM golang:1.10.1-alpine3.7 as builder
MAINTAINER Monax <support@monax.io>

RUN apk add --no-cache --update git bash make

ARG REPO=$GOPATH/src/github.com/hyperledger/burrow
COPY . $REPO
WORKDIR $REPO

# Build purely static binaries
RUN make build

# This will be our base container image
FROM alpine:3.7

ARG REPO=/go/src/github.com/hyperledger/burrow

ENV USER monax
ENV MONAX_PATH /home/$USER/.monax
RUN addgroup -g 101 -S $USER && adduser -S -D -u 1000 $USER $USER
WORKDIR $MONAX_PATH
USER $USER:$USER

# Copy binaries built in previous stage
COPY --from=builder $REPO/bin/* /usr/local/bin/
#RUN chown $USER:$USER /usr/local/bin/burrow*

# Expose ports for 1337:burrow API; 46656:tendermint-peer; 46657:tendermint-rpc
EXPOSE 1337
EXPOSE 46656
EXPOSE 46657

CMD [ "burrow" ]
