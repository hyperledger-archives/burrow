package types

// Defined event input types - these are currently align with EVM types but they technically define a pair/mapping
// of EVM type -> SQL type
const (
	EventFieldTypeInt     = "int"
	EventFieldTypeUInt    = "uint"
	EventFieldTypeAddress = "address"
	EventFieldTypeBytes   = "bytes"
	EventFieldTypeBool    = "bool"
	EventFieldTypeString  = "string"
)
