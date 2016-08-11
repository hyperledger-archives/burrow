# Eris DB

|[![GoDoc](https://godoc.org/github.com/eris-db?status.png)](https://godoc.org/github.com/eris-ltd/eris-db) | Linux |
|---|-------|
| Master | [![Circle CI](https://circleci.com/gh/eris-ltd/eris-db/tree/master.svg?style=svg)](https://circleci.com/gh/eris-ltd/eris-db/tree/master) |
| Develop | [![Circle CI (develop)](https://circleci.com/gh/eris-ltd/eris-db/tree/develop.svg?style=svg)](https://circleci.com/gh/eris-ltd/eris-db/tree/develop) |

Eris DB is Eris' blockchain client. It includes a permissions layer, an implementation of the Ethereum Virtual Machine, and uses Tendermint Consensus. Most functionality is provided by `eris chains`, exposed through [eris-cli](https://github.com/eris-ltd/eris-cli), the entry point for the Eris Platform. 

## Table of Contents

- [Background](#background)
- [Installation](#installation)
- [Usage](#usage)
  - [Security](#security)
- [Contribute](#contribute)
- [License](#license)

## Background

See the [eris-db documentation](https://erisindustries.com/components/erisdb/) for more information.

## Installation

`eris-db` is intended to be used by the `eris chains` command via [eris-cli](https://github.com/eris-ltd/eris-cli). Available commands such as `make | start | stop | logs | inspect | update` are used for chain lifecycle management. 

### For Developers

1. [Install go](https://golang.org/doc/install)
2. Ensure you have `gmp` installed (`sudo apt-get install libgmp3-dev || brew install gmp`)
3. `go get github.com/eris-ltd/eris-db/cmd/erisdb`


To run `erisdb`, just type `$ erisdb /path/to/working/folder`

This will start the node using the provided folder as working dir. If the path is omitted it defaults to `~/.erisdb` 


## Usage

Once the server has started, it will begin syncing up with the network. At that point you may begin using it. The preferred way is through our [javascript api](https://github.com/eris-ltd/eris-db.js), but it is possible to connect directly via HTTP or websocket. The JSON-RPC and web-api reference can be found [here](api)

### Configuration Files

Three files are currently required: 
```
config.toml
genesis.json
priv_validator.json
```
while `server_conf.toml` is optional

### Security

**NOTE**: **CORS** and **TLS** are not yet fully implemented, and cannot be used. CORS is implemented through [gin middleware](https://github.com/tommy351/gin-cors), and TLS through the standard Go http package via the [graceful library](https://github.com/tylerb/graceful).

### server_conf.toml (example)

```
[bind]
address="0.0.0.0"
port=1337
[TLS]
tls=false
cert_path=""
key_path=""
[CORS]
enable=false
allow_origins=[]
allow_credentials=false
allow_methods=[]
allow_headers=[]
expose_headers=[]
max_age=0
[HTTP]
json_rpc_endpoint="/rpc"
[web_socket]
websocket_endpoint="/socketrpc"
max_websocket_sessions=50
read_buffer_size = 4096
write_buffer_size = 4096
[logging]
console_log_level="info"
file_log_level="warn"
log_file=""
```

#### Bind

- `address` is the address.
- `port` is the port number

#### TLS

- `tls` is used to enable/disable TLS
- `cert_path` is the absolute path to the certificate file.
- `key_path` is the absolute path to the key file.

#### CORS

- `enable` is whether or not the CORS middleware should be added at all. **Not implemented:** see above.

#### HTTP

- `json_rpc_endpoint` is the name of the endpoint used for JSON-RPC (2.0) over HTTP.

#### web_socket

- `websocket_endpoint` is the name of the endpoint that is used to establish a websocket connection.
- `max_websocket_connections` is the maximum number of websocket connections that is allowed at the same time.
- `read_buffer_size` is the size of the read buffer for each socket in bytes.
- `read_buffer_size` is the size of the write buffer for each socket in bytes.

#### logging

- `console_log_level` is the logging level used for the console.
- `file_log_level` is the logging level used for the log file.
- `log_file` is the path to the log file. Leaving this empty means file logging will not be used.

The possible log levels are these: `crit`, `error`, `warn`, `info`, `debug`.

The server log level will override the log level set in the Tendermint `config.toml`.

## Contribute

See the [eris platform contributing file here](https://github.com/eris-ltd/coding/blob/master/github/CONTRIBUTING.md).

## License

[GPL-3](LICENSE)
