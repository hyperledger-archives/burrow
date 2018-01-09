FROM quay.io/monax/build:0.16
MAINTAINER Monax <support@monax.io>

# Install monax-keys, a go app for development signing
ENV TARGET monax-keys
ENV REPO $GOPATH/src/github.com/monax/keys

# required for testing; should be removed
RUN apk --no-cache --update add openssl

ADD ./glide.yaml $REPO/
ADD ./glide.lock $REPO/
WORKDIR $REPO
RUN glide install

COPY . $REPO/.
RUN cd $REPO/cmd/$TARGET && \
  go build --ldflags '-extldflags "-static"' -o $INSTALL_BASE/$TARGET
