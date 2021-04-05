# For solc binary
FROM ethereum/solc:0.5.12 as solc-builder
# We use a multistage build to avoid bloating our deployment image with build dependencies
FROM golang:1.15-alpine3.12 as builder

RUN apk add --no-cache --update git bash make musl-dev gcc libc6-compat

ARG REPO=/src/burrow
COPY . $REPO
WORKDIR $REPO

# Build purely static binaries
RUN make build

# This will be our base container image
FROM alpine:3.11

# Variable arguments to populate labels
ARG USER=burrow
ARG INSTALL_BASE=/usr/local/bin

# Fixed labels according to container label-schema
LABEL org.label-schema.schema-version="1.0"
LABEL org.label-schema.name = "Burrow"
LABEL org.label-schema.vendor="Hyperledger Burrow Authors"
LABEL org.label-schema.description="Hyperledger Burrow is a permissioned Ethereum smart-contract blockchain node."
LABEL org.label-schema.license="Apache-2.0"
LABEL org.label-schema.vcs-url="https://github.com/hyperledger/burrow"

# Run burrow as burrow user; not as root user
ENV BURROW_PATH /home/$USER
RUN addgroup -g 101 -S $USER && adduser -S -D -u 1000 $USER $USER
WORKDIR $BURROW_PATH
ADD --chown=$USER /testnet/ $BURROW_PATH

# Copy binaries built in previous stage
COPY --from=builder /src/burrow/bin/burrow $INSTALL_BASE/
COPY --from=builder /src/burrow/bin/burrow-debug $INSTALL_BASE/
COPY --from=solc-builder /usr/bin/solc $INSTALL_BASE/

# Expose ports for 26656:peer; 26658:info; 10997:grpc
EXPOSE 26656
EXPOSE 26658
EXPOSE 10997

USER $USER:$USER
ENTRYPOINT [ "burrow" ]
CMD [ "start" ]
