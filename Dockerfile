# Pull eris/data
FROM eris/data

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

# set the repo and install tendermint
ENV repo /go/src/github.com/eris-ltd/eris-db
ADD . $repo
WORKDIR $repo
RUN cd ./cmd/erisdb && go install
USER eris
ENTRYPOINT ["erisdb"]