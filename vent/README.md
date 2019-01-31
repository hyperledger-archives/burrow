# Vent Component

Vent reads sqlsol specification & abi files, parses their contents and maps column types to corresponding sql types to create or alter database structures.

Then listens to burrow gRPC server and get flowing events based on block range.

Parses, unpacks, decodes data and builds rows to be upserted in matching event tables, rows are upserted in blocks where each block is one commit.

Block id and context info are stored in Log tables in order to resume getting pending blocks.

Given a sqlsol specification 

```json
[
  {
    "TableName" : "EventTest",
    "Filter" : "Log1Text = 'LOGEVENT1'",
    "DeleteFilter": "CRUD_ACTION = 'delete'",
    "Columns"  : {
      "key" : {"name" : "testname", "type": "bytes32", "primary" : true},
      "description": {"name" : "testdescription", "type": "bytes32", "primary" : false, "bytesToString": true}
    }
  },
  {
    "TableName" : "UserAccounts",
    "Filter" : "Log1Text = 'USERACCOUNTS'",
    "Columns"  : {
      "userAddress" : {"name" : "address", "type": "address", "primary" : true},
      "userName": {"name" : "username", "type": "string", "primary" : false}
    }
  }
]

```

Vent builds dictionary, log and event database tables for the defined tables & columns and maps input types to proper sql types.

Database structures are created or altered on the fly based on specifications (just adding new columns is supported).

Abi files can be generated from bin files like so:

```bash
cat *.bin | jq '.Abi[] | select(.type == "event")' > events.abi
```


## Adapters:

Adapters are database implementations, Vent can store data in different rdbms.

In `sqldb/adapters` there's a list of supported adapters (there is also a README.md file in that folder that helps to understand how to implement a new one).

## Setup PostgreSQL Database with Docker:

```bash
# Create postgres container (only once):
docker run --name postgres-local -e POSTGRES_USER=user -e POSTGRES_PASSWORD=pass -e POSTGRES_DB=vent -p 5432:5432 -d postgres:10.4-alpine

# Start postgres container:
docker start postgres-local

# Stop postgres container:
docker stop postgres-local

# Delete postgres container:
docker container rm postgres-local
```

## Run Unit Tests:

```bash
# From the main repo folder:
make test_integration_vent
```

## Run Vent Command:

```bash
# Install vent command:
go install ./vent

# Print command help:
vent --help

# Run vent command with postgres adapter, spec & abi files path, also stores block & tx data:
vent --db-adapter="postgres" --db-url="postgres://user:pass@localhost:5432/vent?sslmode=disable" --db-schema="vent" --grpc-addr="localhost:10997" --http-addr="0.0.0.0:8080" --log-level="debug" --spec-file="<sqlsol specification file path>" --abi-file="<abi file path>" --db-block=true

# Run vent command with sqlite adapter, spec & abi directories path, does not store block & tx data:
vent --db-adapter="sqlite" --db-url="./vent.sqlite" --grpc-addr="localhost:10997" --http-addr="0.0.0.0:8080" --log-level="debug" --spec-dir="<sqlsol specification directory path>" --abi-dir="<abi files directory path>"
```

Configuration Flags:

+ `db-adapter`: (string) Database adapter, 'postgres' or 'sqlite' are fully supported
+ `db-url`: (string) PostgreSQL database URL or SQLite db file path
+ `db-schema`: (string) PostgreSQL database schema or empty for SQLite
+ `http-addr`: (string) Address to bind the HTTP server
+ `grpc-addr`: (string) Address to listen to gRPC Hyperledger Burrow server
+ `log-level`: (string) Logging level (error, warn, info, debug)
+ `spec-file`: (string) SQLSol specification json file (full path)
+ `spec-dir`: (string) Path of a folder to look for SQLSol json specification files
+ `abi-file`: (string) Event Abi specification file full path
+ `abi-dir`: (string) Path of a folder to look for event Abi specification files
+ `db-block`: (boolean) Create block & transaction tables and persist related data (true/false)


NOTES:

One of `spec-file` or `spec-dir` must be provided.
If `spec-dir` is given, vent will search for all `.json` spec files in given directory.

Also one of `abi-file` or `abi-dir` must be provided.
If `abi-dir` is given, vent will search for all `.abi` spec files in given directory.

if `db-block` is set to true (block explorer mode), Block and Transaction tables are created in addition to log and event tables to store block & tx raw info.

It can be checked that vent is connected and ready sending a request to `http://<http-addr>/health` which will return a `200` OK response in case everything's fine.
