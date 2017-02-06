// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

var defaultConfig = `# Copyright 2017 Monax Industries Limited
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License..

# This is a TOML configuration for Eris-DB chains

[chain]

# ChainId is a human-readable name to identify the chain.
# This must correspond to the chain_id defined in the genesis file
# and the assertion here provides a safe-guard on misconfiguring chains.
assert_chain_id = "MyChainId"
# semantic major and minor version
major_version = 0
minor_version = 17
# genesis file, relative path is to eris-db working directory
genesis_file = "genesis.json"


###############################################################################
##
##  consensus
##
###############################################################################

  [chain.consensus]
  # consensus defines the module to use for consensus and
  # this will define the peer-to-peer consensus network;
  # accepted values are "noops", "abci", "tendermint"
  name = "tendermint"
  # version is the major and minor semantic version;
  # the version will be asserted on
  major_version = 0
  minor_version = 8
  # relative path to consensus' module root folder
  relative_root = "tendermint"

###############################################################################
##
##  application manager
##
###############################################################################

  [chain.manager]
  # application manager name defines the module to use for handling
  name = "erismint"
  # version is the major and minor semantic version;
  # the version will be asserted on
  major_version = 0
  minor_version = 17
  # relative path to application manager root folder
  relative_root = "erismint"

################################################################################
################################################################################
##
## Server configurations
##
################################################################################
################################################################################

[servers]

  [servers.bind]
  address = ""
  port = 1337

  [servers.tls]
  tls = false
  cert_path = ""
  key_path = ""

  [servers.cors]
  enable = false
  allow_origins = []
  allow_credentials = false
  allow_methods = []
  allow_headers = []
  expose_headers = []
  max_age = 0

  [servers.http]
  json_rpc_endpoint = "/rpc"

  [servers.websocket]
  endpoint = "/socketrpc"
  max_sessions = 50
  read_buffer_size = 4096
  write_buffer_size = 4096

	[servers.tendermint]
	# Multiple listeners can be separated with a comma
	rpc_local_address = "0.0.0.0:36657"
	endpoint = "/websocket"

################################################################################
################################################################################
##
## Module configurations - dynamically loaded based on chain configuration
##
################################################################################
################################################################################


################################################################################
##
## Tendermint Socket Protocol (abci)
##
## abci expects a tendermint consensus process to run and connect to Eris-DB
##
################################################################################

[abci]
# listener address for accepting tendermint socket protocol connections
listener = "tcp://0.0.0.0:46658"

################################################################################
##yeah we had partial support for that with TMSP
## Tendermint
##
## in-process execution of Tendermint consensus engine
##
################################################################################

[tendermint]
# private validator file is used by tendermint to keep the status
# of the private validator, but also (currently) holds the private key
# for the private vaildator to sign with.  This private key needs to be moved
# out and directly managed by eris-keys
# This file needs to be in the root directory
private_validator_file = "priv_validator.json"

  # Tendermint requires additional configuration parameters.
  # Eris-DB's tendermint consensus module will load [tendermint.configuration]
  # as the configuration for Tendermint.
  # Eris-DB will respect the configurations set in this file where applicable,
  # but reserves the option to override or block conflicting settings.
  [tendermint.configuration]
  # moniker is the name of the node on the tendermint p2p network
  moniker = "anonymous_marmot"
  # seeds lists the peers tendermint can connect to join the network
  seeds = ""
  # fast_sync allows a tendermint node to catch up faster when joining
  # the network.
  # NOTE: Tendermint has reported potential issues with fast_sync enabled.
  # The recommended setting is for keeping it disabled.
  fast_sync = false
  db_backend = "leveldb"
  log_level = "info"
  # node local address
  node_laddr = "0.0.0.0:46656"
  # rpc local address
	# NOTE: value is ignored when run in-process as RPC is
	# handled by [servers.tendermint]
  rpc_laddr = ""
  # proxy application address - used for abci connections,
  # and this port should not be exposed for in-process Tendermint
  proxy_app = "tcp://127.0.0.1:46658"

  # Extended Tendermint configuration settings
  # for reference to Tendermint see https://github.com/tendermint/tendermint/blob/master/config/tendermint/config.go

  # genesis_file = "./data/tendermint/genesis.json"
  # skip_upnp = false
  # addrbook_file = "./data/tendermint/addrbook.json"
  # priv_validator_file = "./data/tendermint/priv_validator.json"
  # db_dir = "./data/tendermint/data"
  # prof_laddr = ""
  # revision_file = "./data/tendermint/revision"
  # cswal = "./data/tendermint/data/cswal"
  # cswal_light = false

  # block_size = 10000
  # disable_data_hash = false
  # timeout_propose = 3000
  # timeout_propose_delta = 500
  # timeout_prevote = 1000
  # timeout_prevote_delta = 500
  # timeout_precommit = 1000
  # timeout_precommit_delta = 500
  # timeout_commit = 1000
  # mempool_recheck = true
  # mempool_recheck_empty = true
  # mempool_broadcast = true

################################################################################
##
## Eris-Mint
## version 0.17.0
##
## The original Ethereum virtual machine with IAVL merkle trees
## and tendermint/go-wire encoding
##
################################################################################

[erismint]
# Database backend to use for ErisMint state database.
db_backend = "leveldb"
`
