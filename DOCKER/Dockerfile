# Pull base image.
FROM eris/base
MAINTAINER Eris Industries <support@erisindustries.com>

# Set the env variables to non-interactive
ENV DEBIAN_FRONTEND noninteractive
ENV DEBIAN_PRIORITY critical
ENV DEBCONF_NOWARNINGS yes
ENV TERM linux
RUN echo 'debconf debconf/frontend select Noninteractive' | debconf-set-selections

# grab deps (gmp)
RUN apt-get update && \
  apt-get install -y --no-install-recommends \
    libgmp3-dev && \
  rm -rf /var/lib/apt/lists/*

# install the wrapper/start script
COPY DOCKER/start.sh /usr/local/bin/erisdb-wrapper

# set the repo and install tendermint
ENV REPO github.com/tendermint/tendermint
ENV BRANCH f2f437a328cb3e277bc4ac74e4a1180f1c28efbc
# note, we want to lock in the same commit number
#  as eris-db has vendored.
RUN mkdir --parents $GOPATH/src/$REPO
WORKDIR $GOPATH/src/$REPO
RUN git clone --quiet https://$REPO . && \
  git checkout --quiet $BRANCH && \
  go build -o /usr/local/bin/tendermint ./cmd/tendermint

# set the repo and install erisdb
ENV REPO $GOPATH/src/github.com/eris-ltd/eris-db
COPY . $REPO
WORKDIR $REPO
RUN go build -o /usr/local/bin/erisdb ./cmd/erisdb && \
  go build -o /usr/local/bin/erisdbss ./cmd/erisdbss

# set the repo and install mint-client
ENV REPO github.com/eris-ltd/mint-client
ENV BRANCH master
RUN mkdir --parents $GOPATH/src/$REPO
WORKDIR $GOPATH/src/$REPO
RUN git clone --quiet https://$REPO . && \
  git checkout --quiet $BRANCH && \
  go install ./... && \
  mv $GOPATH/bin/mint* /usr/local/bin && \
  mv ./mint-client /usr/local/bin/

# cleanup
RUN rm -rf $GOPATH/src/* && \
  unset REPO && \
  unset BRANCH

# install chain manager scripts and default mint config
ENV ECM_PATH /usr/local/lib/ecm/
RUN mkdir -p $ECM_PATH
COPY ./DOCKER/chain_* $ECM_PATH
COPY ./DOCKER/config.toml  $ECM_PATH

# persist data, set user
VOLUME /home/$USER/.eris
WORKDIR /home/$USER/.eris
USER $USER
RUN mkdir --parents /home/$USER/.eris/blockchains/tendermint
ENV TMROOT /home/$USER/.eris/blockchains/tendermint

# run tendermint
# really should be entrypoint but broken in API 1.19
# ports: 1337:eris-db API; 46656:mint-peer; 46657:mint-rpc
EXPOSE 1337
EXPOSE 46656
EXPOSE 46657
CMD ["erisdb-wrapper"]
