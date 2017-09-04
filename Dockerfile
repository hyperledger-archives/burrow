# We use a multistage build to avoid bloating our deployment image with build dependencies
FROM golang:1.9.0-alpine3.6 as builder
MAINTAINER Monax <support@monax.io>

RUN apk add --no-cache --update git
RUN go get github.com/Masterminds/glide

ENV REPO $GOPATH/src/github.com/hyperledger/burrow
COPY . $REPO
WORKDIR $REPO
RUN glide install

# Build purely static binaries
RUN go build --ldflags '-extldflags "-static"' -o bin/burrow ./cmd/burrow
RUN go build --ldflags '-extldflags "-static"' -o bin/burrow-client ./client/cmd/burrow-client

# This will be our base container image
FROM alpine:3.6

# There does not appear to be a way to share environment variables between stages
ENV REPO /go/src/github.com/hyperledger/burrow

ENV USER monax
ENV MONAX_PATH /home/$USER/.monax
RUN addgroup -g 101 -S $USER && adduser -S -D -u 1000 $USER $USER
WORKDIR $MONAX_PATH
USER $USER:$USER

# Copy binaries built in previous stage
COPY --from=builder $REPO/bin/* /usr/local/bin/

# Expose ports for 1337:burrow API; 46656:tendermint-peer; 46657:tendermint-rpc
EXPOSE 1337
EXPOSE 46656
EXPOSE 46657

CMD [ "burrow", "serve" ]
