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
	&& go build \
	&& cp eris-db $INSTALL_BASE/eris-db

#-----------------------------------------------------------------------------
# install mint-client [to be deprecated]

ENV ERIS_DB_MINT_REPO github.com/eris-ltd/mint-client
ENV ERIS_DB_MINT_SRC_PATH $GOPATH/src/$ERIS_DB_MINT_REPO

WORKDIR $ERIS_DB_MINT_SRC_PATH

RUN git clone --quiet https://$ERIS_DB_MINT_REPO . \
	&& git checkout --quiet master \
	&& go build -o $INSTALL_BASE/mintx ./mintx \
	&& go build -o $INSTALL_BASE/mintconfig ./mintconfig 
	# restrict build targets for re-evaluation
	# && go build -o $INSTALL_BASE/mintdump ./mintdump \
	# && go build -o $INSTALL_BASE/mintperms ./mintperms \
	# && go build -o $INSTALL_BASE/mintunsafe ./mintunsafe \
	# && go build -o $INSTALL_BASE/mintkey ./mintkey \
	# && go build -o $INSTALL_BASE/mintgen ./mintgen \
	# && go build -o $INSTALL_BASE/mintsync ./mintsync \

#WORKDIR $ERIS
