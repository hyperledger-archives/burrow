# Pull base image.
FROM quay.io/eris/build
MAINTAINER Eris Industries <support@erisindustries.com>


#-----------------------------------------------------------------------------
# install eris-db
#
ENV GO15VENDOREXPERIMENT 1

# install glide for dependency management
RUN	go get github.com/Masterminds/glide

# set the repo and install erisdb
ENV REPO $GOPATH/src/github.com/eris-ltd/eris-db
WORKDIR $REPO
COPY . $REPO
RUN glide install
RUN cd ./cmd/eris-db && go build -o /usr/local/bin/eris-db && cd ../..

#-----------------------------------------------------------------------------
# install mint-client tools

# set the repo and install mint-client
ENV REPO github.com/eris-ltd/mint-client
ENV BRANCH master
RUN mkdir -p $GOPATH/src/$REPO
WORKDIR $GOPATH/src/$REPO
RUN git clone --quiet https://$REPO . && \
  git checkout --quiet $BRANCH && \
  go install ./... && \
  mv $GOPATH/bin/mint* /usr/local/bin && \
  mv ./mint-client /usr/local/bin/

#-----------------------------------------------------------------------------
# cleanup

RUN rm -f $GOPATH/bin/glide
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
WORKDIR /home/$USER/.eris
VOLUME /home/$USER/.eris
USER $USER
RUN mkdir -p /home/$USER/.eris/chains/tendermint
ENV CHAINROOT /home/$USER/.eris/chains/tendermint

# run eris-db
# ports: 1337:eris-db API; 46656:tendermint-peer; 46657:tendermint-rpc
EXPOSE 1337
EXPOSE 46656
EXPOSE 46657
ENTRYPOINT ["eris-db", "serve"]
# CMD["--work-dir $CHAINROOT"]
