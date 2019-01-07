package types

// SQLErrorType stores generic SQL error types
type SQLErrorType int

// generic SQL error types
const (
	SQLErrorTypeDuplicatedSchema SQLErrorType = iota
	SQLErrorTypeDuplicatedColumn
	SQLErrorTypeDuplicatedTable
	SQLErrorTypeInvalidType
	SQLErrorTypeUndefinedTable
	SQLErrorTypeUndefinedColumn
	SQLErrorTypeGeneric
)
