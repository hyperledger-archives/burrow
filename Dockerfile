FROM quay.io/monax/build:0.16
MAINTAINER Monax <support@monax.io>

ENV TARGET burrow
ENV REPO $GOPATH/src/github.com/hyperledger/$TARGET

ADD ./glide.yaml $REPO/
ADD ./glide.lock $REPO/
WORKDIR $REPO
RUN glide install

COPY . $REPO/.
RUN cd $REPO/cmd/$TARGET && \
  go build --ldflags '-extldflags "-static"' -o $INSTALL_BASE/$TARGET

# build customizations start here
RUN cd $REPO/client/cmd/burrow-client && \
  go build --ldflags '-extldflags "-static"' -o $INSTALL_BASE/burrow-client
