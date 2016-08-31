#!/usr/bin/env bash
# -----------------------------------------------------------------------------
# PURPOSE

# This script will setup docker.

# **NOTE** -- This script is used by Eris to provision test box backends which
# we need to be on a specific version of docker. It will likely be unuseful to
# you.

# If you are looking for a quick and easy way to set up eris on a cloud machin
# please see https://github.com/eris-ltd/common/cloud/chains/setup/setup.sh

# If $DOCKER_VERSION is set then the host will use that.

# -----------------------------------------------------------------------------
# LICENSE

# The MIT License (MIT)
# Copyright (c) 2016-Present Eris Industries, Ltd.

# Permission is hereby granted, free of charge, to any person obtaining a copy
# of this software and associated documentation files (the "Software"), to deal
# in the Software without restriction, including without limitation the rights
# to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
# copies of the Software, and to permit persons to whom the Software is
# furnished to do so, subject to the following conditions:

# The above copyright notice and this permission notice shall be included in
# all copies or substantial portions of the Software.

# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
# AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
# LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
# FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
# IN THE SOFTWARE.

# -----------------------------------------------------------------------------
# REQUIREMENTS

# Ubuntu

# -----------------------------------------------------------------------------
# USAGE

# docker.sh

# -----------------------------------------------------------------------------
# Set defaults

default_docker="1.9.0"

# -----------------------------------------------------------------------------
# Check Ubuntu

#read -p "This script only works on Ubuntu (and does no checking). It may work on some debians. Do you wish to proceed? (y/n) " -n 1 -r
#echo
#if [[ ! $REPLY =~ ^[Yy]$ ]]
#then
#  echo "OK. Not doing anything. Bye."
#  exit 1
#fi
#echo "You confirmed you are on Ubuntu (or waived compatibility)."

# ----------------------------------------------------------------------------
# Check sudo

if [[ "$USER" != "root" ]]
then
  echo "OK. Not doing anything. Bye."
  exit 1
fi
echo "Privileges confirmed."

# ----------------------------------------------------------------------------
# Check Docker Version to Install

if [ -z "$DOCKER_VERSION" ]
then
  echo "You do not have the \$DOCKER_VERSION set. Trying via hostname (an Eris paradigm)."
  export DOCKER_VERSION=$(hostname | cut -d'-' -f4)
  if [[ "$DOCKER_VERSION" == `hostname` ]]
  then
    read -p "I cannot find the Docker Version to Install. You can rerun me with \$DOCKER_VERSION set or use the defaults. Would you like the defaults? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]
    then
      export DOCKER_VERSION="$default_docker"
    fi
  fi
fi
echo "Will install Docker for Version: $DOCKER_VERSION"
echo
echo

# ---------------------------------------------------------------------------
# Install Docker Version

echo
wget -qO- https://get.docker.io/gpg | apt-key add -
sh -c "echo deb https://get.docker.io/ubuntu docker main > /etc/apt/sources.list.d/docker.list"
apt-get update -qq
apt-get install -qqy docker-engine
service docker stop
docker_place=$(which docker)
rm "$docker_place"
curl -sSL --ssl-req -o "$docker_place" https://get.docker.com/builds/Linux/x86_64/docker-$DOCKER_VERSION
chmod 755 "$docker_place"
echo
echo "Docker installed"

# ---------------------------------------------------------------------------
# Restart Docker

echo "Restarting Newly Installed Docker"
echo
service docker start
sleep 3 # boot time

# ---------------------------------------------------------------------------
# Check User Needs to be Added to Docker group

echo
echo
usermod -a -G docker $USER

# ---------------------------------------------------------------------------
# Cleanup

echo
echo
echo "All set"
echo
echo
