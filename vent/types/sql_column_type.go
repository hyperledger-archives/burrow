package types

// SQLColumnType to store generic SQL column types
type SQLColumnType int

// generic SQL column types
const (
	SQLColumnTypeBool SQLColumnType = iota
	SQLColumnTypeByteA
	SQLColumnTypeInt
	SQLColumnTypeSerial
	SQLColumnTypeText
	SQLColumnTypeVarchar
	SQLColumnTypeTimeStamp
	SQLColumnTypeNumeric
	SQLColumnTypeJSON
	SQLColumnTypeBigInt
)

func (ct SQLColumnType) String() string {
	switch ct {
	case SQLColumnTypeBool:
		return "bool"
	case SQLColumnTypeByteA:
		return "bytea"
	case SQLColumnTypeInt:
		return "int"
	case SQLColumnTypeSerial:
		return "serial"
	case SQLColumnTypeText:
		return "text"
	case SQLColumnTypeVarchar:
		return "varchar"
	case SQLColumnTypeTimeStamp:
		return "timestamp"
	case SQLColumnTypeNumeric:
		return "numeric"
	case SQLColumnTypeJSON:
		return "json"
	case SQLColumnTypeBigInt:
		return "bigint"
	}
	return "unknown SQL type"
}

// IsNumeric determines if an sqlColumnType is numeric
func (ct SQLColumnType) IsNumeric() bool {
	return ct == SQLColumnTypeInt || ct == SQLColumnTypeSerial || ct == SQLColumnTypeNumeric || ct == SQLColumnTypeBigInt
}
