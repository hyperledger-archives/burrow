# Pull base image.
FROM eris/base
MAINTAINER Eris Industries <support@erisindustries.com>

#-----------------------------------------------------------------------------
# dependencies
RUN apt-get update && \
  apt-get install -y --no-install-recommends \
    libgmp3-dev jq && \
  rm -rf /var/lib/apt/lists/*

#-----------------------------------------------------------------------------
# install tendermint

# set the repo and install tendermint -- should track vendored commit
ENV REPO_ALIAS github.com/tendermint/tendermint
ENV REPO github.com/eris-ltd/tendermint
#ENV COMMIT 4ee387d0770ed379e2d524f7077938517b38cd7c
ENV COMMIT working
# note, we want to lock in the same commit number
#  as eris-db has vendored.
RUN mkdir --parents $GOPATH/src/$REPO
WORKDIR $GOPATH/src/$REPO_ALIAS
RUN git clone --quiet https://$REPO . && \
  git checkout --quiet $COMMIT && \
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
ENV BRANCH develop
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
  unset COMMIT && \
  unset BRANCH

#-----------------------------------------------------------------------------
# start script

# install the wrapper/start script
COPY DOCKER/start.sh /usr/local/bin/erisdb-wrapper

#-----------------------------------------------------------------------------
# chain manager scripts and default mint config

ENV ECM_PATH /usr/local/lib/ecm
RUN mkdir -p $ECM_PATH
COPY DOCKER/chain_* $ECM_PATH/
COPY DOCKER/config.toml $ECM_PATH/
COPY DOCKER/version.sh $ECM_PATH/

#-----------------------------------------------------------------------------
# permission the directories
RUN chown -R eris /usr/local/bin
RUN chown -R eris $ECM_PATH

#-----------------------------------------------------------------------------
# root dir

# persist data, set user
VOLUME /home/$USER/.eris
WORKDIR /home/$USER/.eris
USER $USER
RUN mkdir --parents /home/$USER/.eris/blockchains/tendermint
ENV TMROOT /home/$USER/.eris/blockchains/tendermint

# run tendermint
# ports: 1337:eris-db API; 46656:mint-peer; 46657:mint-rpc
EXPOSE 1337
EXPOSE 46656
EXPOSE 46657
ENTRYPOINT ["erisdb-wrapper"]
