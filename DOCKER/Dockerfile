# Pull base image.
FROM eris/base
MAINTAINER Eris Industries <support@erisindustries.com>

#-----------------------------------------------------------------------------
# dependencies
RUN apt-get update && \
  apt-get install -y --no-install-recommends \
    libgmp3-dev && \
  rm -rf /var/lib/apt/lists/*

#-----------------------------------------------------------------------------
# start script

# install the wrapper/start script
COPY DOCKER/start.sh /usr/local/bin/erisdb-wrapper

#-----------------------------------------------------------------------------
# install tendermint

# set the repo and install tendermint
ENV REPO github.com/tendermint/tendermint
ENV BRANCH 4ee387d0770ed379e2d524f7077938517b38cd7c
# note, we want to lock in the same commit number
#  as eris-db has vendored.
RUN mkdir --parents $GOPATH/src/$REPO
WORKDIR $GOPATH/src/$REPO
RUN git clone --quiet https://$REPO . && \
  git checkout --quiet $BRANCH && \
  go build -o /usr/local/bin/tendermint ./cmd/tendermint

#-----------------------------------------------------------------------------
# install erisdb

# set the repo and install erisdb
ENV REPO $GOPATH/src/github.com/eris-ltd/eris-db
COPY . $REPO
WORKDIR $REPO
RUN cd ./cmd/erisdb && go build -o /usr/local/bin/erisdb && \
  cd ../erisdbss && go build -o /usr/local/bin/erisdbss && cd ../..

#-----------------------------------------------------------------------------
# install mint-client tools

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

#-----------------------------------------------------------------------------
# cleanup

RUN rm -rf $GOPATH/src/* && \
  unset REPO && \
  unset BRANCH

#-----------------------------------------------------------------------------
# chain manager scripts and default mint config

ENV ECM_PATH /usr/local/lib/ecm/
RUN mkdir -p $ECM_PATH
COPY DOCKER/chain_* $ECM_PATH
COPY DOCKER/config.toml $ECM_PATH

#-----------------------------------------------------------------------------
# root dir

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
