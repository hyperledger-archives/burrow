# Vent - SQL mapping layer


Vent reads specification files called 'projections', parses their contents, and maps EVM LOG event fields to corresponding SQL columns to create or alter database structures. 
It listens for a stream of block events from Burrow's GRPC service then parses, unpacks, decodes event data, and builds rows to be upserted in matching event tables, rows are 
upserted atomically in a single database transaction per block.

There are two modes of operation: view mode and log mode. In view mode a primary key is used to locate the row in a table which should be updated (if exists) or inserted 
(if it does not exist). If the event contains a field matching the optional `DeleteMarkerField` then the row will instead be deleted. As such in view mode Vent can map a 
stream of EVM LOG events to a CRUD-style table - a view over entities as defined by the choice primary key. Alternatively if no primary keys are specified for a projection 
Vent operates in log mode where all matched events are inserted - and so log mode operates as an append-only log. Note there is no explicit setting for mode - it depends on 
the presence or absence of a `"Primary": true` entry in one of the `FieldMappings` of a projection (see below for an example).

Vent writes each block of updates atomically and is guaranteed to be crash tolerant. If the Vent process is killed it will resume at the last written height. Burrow stores all 
previous events in its state so even if you delete the Vent database it can be regenerated deterministically. This feature being a core feature of Vent.

There is a [presentation on vent here](https://competent-yalow-f210f7.netlify.app).


## Projections
A projection is the name  given to the configuration files that Vent uses to interpret EVM events as updates or deletion from SQL tables. They provide an object relational mapping 
between Solidity events and SQL tables.

Given a projection, like the following:

```json
[
  {
    "TableName" : "EventTest",
    "Filter" : "Log1Text = 'I am LOG1'",
    "DeleteMarkerField": "__DELETE__",
    "FieldMappings"  : [
      {
        "Field": "key", 
        "ColumnName" : "testname", 
        "Type": "bytes32", 
        "Primary" : true
      },
      {
        "Field": "description", 
        "ColumnName" : "testdescription", 
        "Type": "bytes32", 
        "BytesToString": true
      }
    ]
  }
]
```

And a solidity contract like:

```solidity
pragma solidity ^0.4.25;

contract EventEmitter {
    event UpdateEvent(
        // The first indexed field will appear as the the LOG1 topic - we can use it like a namespace
        bytes32 indexed IAmLog1,
        // Our primary key in our projection above
        bytes32 key,
        // Some 'mutable' text - we can update this by emitting an UpdateEvent with the same key but a new description
        bytes32 description
    );

    event DeletionEvent(
        bytes32 indexed IAmLog1,
        bytes32 key,
        // This marker field can be of any type - it is purely matched on name - if an event contains a field with the
        // the specified marker field it is interpreted as an instruction to delete the row corresponding to key
        bool __DELETE__
    );
    
    function update() external {
        // Update or inserts 'key0001' row
        emit UpdateEvent("I am LOG1", "key0001", "some description");
    }

    function update2() external {
        // Update or inserts 'key0001' row
        emit UpdateEvent("I am LOG1", "key0001", "a different description");
    }

    function remove() external {
        // Removes 'key0001' row
        emit DeletionEvent("I am LOG1", "key0001", true);
    }
}
```

We can maintain a view-mode table that feels like that of a ordinary CRUD app though it is backed by a stream of events coming from our Solidity contracts.

Burrow can also emit a JSONSchema for the projection file format with `burrow vent schema`. You can use this to validate your projections using any of the 
[JSONSchema](https://json-schema.org/) tooling.

### Projection specification
A projection file is defined as a JSON array of `EventClass` objections. Each `EventClass` specifies a class of events that should be consumed (specified via a filter) 
in order to generate a SQL table. An `EventClass` holds `FieldMappings` that specify how to map the event fields of a matched EVM event to a destination column 
(identified by `ColumnName`) of the destination table (identified by `TableName`)

#### EventClass
| Field | Type | Required? | Description |
|-------|------|-----------|-------------|
| `TableName` | String | Required | The case-sensitive name of the destination SQL table for the `EventClass`|
| `Filter` | String | Required | A filter to be applied to EVM Log events using the [available tags](../../protobuf/rpcevents.proto) written according to the event [query.peg](../../event/query/query.peg) grammar |
| `FieldMappings` | array of `FieldMapping` | Required | Mappings between EVM event fields and columns see table below |
| `DeleteMarkerField` | String | Optional | Field name of an event field that when present in a matched event indicates the event should result on a deletion of a row (matched on the primary keys of that row) rather than the default upsert action |

#### FieldMapping
| Field | Type | Required? | Description |
|-------|------|-----------|-------------|
| `Field` | String | Required | EVM field name to match exactly when creating a SQL upsert/delete |
| `Type` | String | Required | EVM type of the field (which also dictates the SQL type that will be used for table definition) |
| `ColumnName` | String | Required | The destination SQL column for the mapped value |
| `Primary` | Boolean | Optional | Whether this SQL column should be part of the primary key |
| `BytesToString` | Boolean | Optional | When type is `bytes<N>` (for some N) indicates that the value should be interpreted as (converted to) a string  |
| `Notify` | array of String | Optional | A list of notification channels on which a payload should be sent containing the value of this column when it is updated or deleted. The payload on a particular channel will be the JSON object containing all column/value pairs for which the notification channel is a member of this notify array (see [triggers](#triggers) below) |

Vent builds dictionary, log and event database tables for the defined tables & columns and maps input types to proper sql types.

Database structures are created or altered on the fly based on specifications (just adding new columns is supported).

Abi files can be generated from bin files like so:

```bash
cat *.bin | jq '.Abi[] | select(.type == "event")' > events.abi
```

## Adapters:

Adapters are database implementations, Vent can store data in different rdbms.

In `sqldb/adapters` there's a list of supported adapters (there is also a README.md file in that folder that helps to understand how to implement a new one).

### <a name="triggers"></a>Notification Triggers
Notification triggers are configured with the `Notify` array of a `FieldMapping`. In a supported database (currently only postrges) they allow you to specify a set of 
channels on which to notify when a column changes. By including a channel in the `Notify` the column is added to the set of columns for which that channel should receive 
a notification payload. For example if we have the following spec:

```json
[  
  {
    "TableName" : "UserAccounts",
    "Filter" : "Log1Text = 'USERACCOUNTS'",
    "FieldMappings"  : [
      {"Field": "userAddress", "ColumnName" : "address", "Type": "address", "Notify": ["user", "address"]},
      {"Field": "userName", "ColumnName" : "username", "Type": "string", "Notify":  ["user"]}
    ]
  }
 ]
```

Then Vent will record a mapping `user -> username, address` and `address -> address` where the left hand side is the notification channel and the right hand side the columns 
included in the payload on that channel.

For each of these mappings a notification trigger function is defined and attached as a trigger for the table to run after an insert, update, or delete. This function calls 
`pg_notify` (in the case of postgres, the only database for which we support notifications - this is non-standard and we may use a different mechanism in other databases if present). 
These notification can be consumed by any client connected to the postgres database with `LISTEN <channel>;`, see [Postgres NOTIFY documentation](https://www.postgresql.org/docs/11/sql-notify.html).

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
# Install burrow (from root of repo):
make install

# Print command help:
burrow vent --help

# Run vent command with postgres adapter, spec & abi files path, also stores block & tx data:
burrow vent start --db-adapter="postgres" --db-url="postgres://user:pass@localhost:5432/vent?sslmode=disable" --db-schema="vent" --grpc-addr="localhost:10997" --http-addr="0.0.0.0:8080" --log-level="debug" --spec="<sqlsol specification file path>" --abi="<abi file path>" --db-block=true

# Run vent command with sqlite adapter, spec & abi directories path, does not store block & tx data:
burrow vent start --db-adapter="sqlite" --db-url="./vent.sqlite" --grpc-addr="localhost:10997" --http-addr="0.0.0.0:8080" --log-level="debug" --spec="<sqlsol specification directory path>" --abi="<abi files directory path>"
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
