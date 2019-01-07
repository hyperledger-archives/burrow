# Vent Adapters

Vent adapters are relational dbms that can be used to store event & log data.

## Supported adapters:

+ PostgreSQL v9 (and above) is fully supported.
+ SQLite v3 (and above) is fully supported.

## Considerations for adding new adapters:

Each adapter must be in a separate file with the name `<dbms>_adapter.go` and must implement given interface methods described in `db_adapter.go`.

This is all that is needed to add a new rdbms adapter, in addition to importing proper database driver.

Provided implementations are included in `postgres_adapter.go` and `sqlite_adapter.go`.
