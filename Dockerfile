# Pull base image.
FROM quay.io/eris/build
MAINTAINER Eris Industries <support@erisindustries.com>

# Expose ports for 1337:eris-db API; 46656:tendermint-peer; 46657:tendermint-rpc
EXPOSE 1337
EXPOSE 46656
EXPOSE 46657

#-----------------------------------------------------------------------------
# install eris-db

# set the source code path and copy the repository in
ENV ERIS_DB_SRC_PATH $GOPATH/src/github.com/eris-ltd/eris-db
COPY . $ERIS_DB_SRC_PATH

# fetch and install eris-db and its dependencies
	# install glide for dependency management
RUN go get github.com/Masterminds/glide \
	# build the main eris-db target
	&& cd $ERIS_DB_SRC_PATH/cmd/eris-db \
	# install dependencies for eris-db with glide
	&& glide install \
	&& go build \
	&& cp eris-db $INSTALL_BASE/eris-db \
	# copy the start script for eris-db \
	&& cp $ERIS_DB_SRC_PATH/bin/start_eris_db $INSTALL_BASE/erisdb-wrapper \
	&& chmod 755 $INSTALL_BASE/erisdb-wrapper

#-----------------------------------------------------------------------------
# install mint-client [to be deprecated]

ENV ERIS_DB_MINT_REPO github.com/eris-ltd/mint-client
ENV ERIS_DB_MINT_SRC_PATH $GOPATH/src/$ERIS_DB_MINT_REPO

WORKDIR $ERIS_DB_MINT_SRC_PATH

RUN git clone --quiet https://$ERIS_DB_MINT_REPO . \
	&& git checkout --quiet master \
	&& go build -o $INSTALL_BASE/mintx ./mintx \
	&& go build -o $INSTALL_BASE/mintconfig ./mintconfig \
	&& go build -o $INSTALL_BASE/mintkey ./mintkey
	# restrict build targets for re-evaluation
	# && go build -o $INSTALL_BASE/mintdump ./mintdump \
	# && go build -o $INSTALL_BASE/mintperms ./mintperms \
	# && go build -o $INSTALL_BASE/mintunsafe ./mintunsafe \
	# && go build -o $INSTALL_BASE/mintgen ./mintgen \
	# && go build -o $INSTALL_BASE/mintsync ./mintsync

#-----------------------------------------------------------------------------
# clean up [build container needs to be separated from shipped container]

RUN unset ERIS_DB_SRC_PATH \
	&& unset ERIS_DB_MINT_SRC_PATH \
	&& apk del --purge go git musl-dev \
	&& rm -rf $GOPATH

# mount the data container on the eris directory
VOLUME $ERIS

WORKDIR $ERIS

CMD "erisdb-wrapper"