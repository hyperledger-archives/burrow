# We use a multistage build to avoid bloating our deployment image with build dependencies
FROM golang:1.10.3-alpine3.8 as builder

RUN apk add --no-cache --update git bash make

ARG REPO=$GOPATH/src/github.com/hyperledger/burrow
COPY . $REPO
WORKDIR $REPO

# Build purely static binaries
RUN make build

# This will be our base container image
FROM alpine:3.8

# Variable arguments to populate labels
ARG VERSION
ARG VCS_REF=master
ARG BUILD_DATE

# Fixed labels according to container label-schema
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.name = "Burrow"
LABEL org.label-schema.vendor="Hyperledger Burrow Authors"
LABEL org.label-schema.description="Hyperledger Burrow is a permissioned Ethereum smart-contract blockchain node."
LABEL org.label-schema.license="Apache-2.0"
LABEL org.label-schema.version=$VERSION
LABEL org.label-schema.vcs-url="https://github.com/hyperledger/burrow"
LABEL org.label-schema.vcs-ref=$VCS_REF
LABEL org.label-schema.build-date=$BUILD_DATE

# Run burrow as burrow user; not as root user
ENV USER burrow
ENV BURROW_PATH /home/$USER
RUN addgroup -g 101 -S $USER && adduser -S -D -u 1000 $USER $USER
WORKDIR $BURROW_PATH

# Copy binaries built in previous stage
COPY --from=builder /go/src/github.com/hyperledger/burrow/bin/burrow /usr/local/bin/

# Expose ports for 26656:peer; 26658:info; 10997:grpc
EXPOSE 26656
EXPOSE 26658
EXPOSE 10997

USER $USER:$USER
ENTRYPOINT [ "burrow" ]
