package sqlsol

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/hyperledger/burrow/vent/types"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// Projection contains EventTable, Event & Abi specifications
type Projection struct {
	Tables    types.EventTables
	EventSpec types.EventSpec
}

// NewProjectionFromBytes creates a Projection from a stream of bytes
func NewProjectionFromBytes(bs []byte) (*Projection, error) {
	eventSpec := types.EventSpec{}

	err := ValidateJSONEventSpec(bs)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bs, &eventSpec)
	if err != nil {
		return nil, errors.Wrap(err, "Error unmarshalling eventSpec")
	}

	return NewProjectionFromEventSpec(eventSpec)
}

// NewProjectionFromFolder creates a Projection from a folder containing spec files
func NewProjectionFromFolder(specFileOrDirs ...string) (*Projection, error) {
	eventSpec := types.EventSpec{}

	const errHeader = "NewProjectionFromFolder():"

	for _, dir := range specFileOrDirs {
		err := filepath.Walk(dir, func(path string, _ os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("error walking event spec files location '%s': %v", dir, err)
			}
			if filepath.Ext(path) == ".json" {
				bs, err := readFile(path)
				if err != nil {
					return fmt.Errorf("error reading spec file '%s': %v", path, err)
				}

				err = ValidateJSONEventSpec(bs)
				if err != nil {
					return fmt.Errorf("could not validate spec file '%s': %v", path, err)
				}

				fileEventSpec := types.EventSpec{}
				err = json.Unmarshal(bs, &fileEventSpec)
				if err != nil {
					return fmt.Errorf("error reading spec file '%s': %v", path, err)
				}

				eventSpec = append(eventSpec, fileEventSpec...)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("%s %v", errHeader, err)
		}
	}

	return NewProjectionFromEventSpec(eventSpec)
}

// NewProjectionFromEventSpec receives a sqlsol event specification
// and returns a pointer to a filled projection structure
// that contains event types mapped to SQL column types
// and Event tables structures with table and columns info
func NewProjectionFromEventSpec(eventSpec types.EventSpec) (*Projection, error) {
	// builds abi information from specification
	tables := make(types.EventTables)

	// obtain global field mappings to add to table definitions
	globalFieldMappings := getGlobalFieldMappings()

	for _, eventClass := range eventSpec {
		// validate json structure
		if err := eventClass.Validate(); err != nil {
			return nil, fmt.Errorf("validation error on %v: %v", eventClass, err)
		}

		// build columns mapping
		var columns []*types.SQLTableColumn
		channels := make(map[string][]string)

		// Add the global mappings
		eventClass.FieldMappings = append(globalFieldMappings, eventClass.FieldMappings...)

		i := 0
		for _, mapping := range eventClass.FieldMappings {
			sqlType, sqlTypeLength, err := getSQLType(mapping.Type, mapping.BytesToString)
			if err != nil {
				return nil, err
			}

			i++

			// Update channels broadcast payload subsets with this column
			for _, channel := range mapping.Notify {
				channels[channel] = append(channels[channel], mapping.ColumnName)
			}

			columns = append(columns, &types.SQLTableColumn{
				Name:    mapping.ColumnName,
				Type:    sqlType,
				Primary: mapping.Primary,
				Length:  sqlTypeLength,
			})
		}

		// Allow for compatible composition of tables
		var err error
		tables[eventClass.TableName], err = mergeTables(tables[eventClass.TableName],
			&types.SQLTable{
				Name:           eventClass.TableName,
				NotifyChannels: channels,
				Columns:        columns,
			})
		if err != nil {
			return nil, err
		}

	}

	// check if there are duplicated duplicated column names (for a given table)
	colName := make(map[string]int)

	for _, table := range tables {
		for _, column := range table.Columns {
			colName[table.Name+column.Name]++
			if colName[table.Name+column.Name] > 1 {
				return nil, fmt.Errorf("duplicated column name: '%s' in table '%s'", column.Name, table.Name)
			}
		}
	}

	return &Projection{
		Tables:    tables,
		EventSpec: eventSpec,
	}, nil
}

// Get the column for a particular table and column name
func (p *Projection) GetColumn(tableName, columnName string) (*types.SQLTableColumn, error) {
	if table, ok := p.Tables[tableName]; ok {
		column := table.GetColumn(columnName)
		if column == nil {
			return nil, fmt.Errorf("GetColumn: table '%s' has no column '%s'",
				tableName, columnName)
		}
		return column, nil
	}

	return nil, fmt.Errorf("GetColumn: table does not exist projection: %s ", tableName)
}

func ValidateJSONEventSpec(bs []byte) error {
	schemaLoader := gojsonschema.NewGoLoader(types.EventSpecSchema())
	specLoader := gojsonschema.NewBytesLoader(bs)
	result, err := gojsonschema.Validate(schemaLoader, specLoader)
	if err != nil {
		return fmt.Errorf("could not validate using JSONSchema: %v", err)
	}

	if !result.Valid() {
		errs := make([]string, len(result.Errors()))
		for i, err := range result.Errors() {
			errs[i] = err.String()
		}
		return fmt.Errorf("EventSpec failed JSONSchema validation:\n%s", strings.Join(errs, "\n"))
	}
	return nil
}

// readFile opens a given file and reads it contents into a stream of bytes
func readFile(file string) ([]byte, error) {
	theFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer theFile.Close()

	byteValue, err := ioutil.ReadAll(theFile)
	if err != nil {
		return nil, err
	}

	return byteValue, nil
}

// getSQLType maps event input types with corresponding SQL column types
// takes into account related solidity types info and element indexed or hashed
func getSQLType(evmSignature string, bytesToString bool) (types.SQLColumnType, int, error) {
	evmSignature = strings.ToLower(evmSignature)
	re := regexp.MustCompile("[0-9]+")
	typeSize, _ := strconv.Atoi(re.FindString(evmSignature))

	switch {
	// solidity address => sql varchar
	case evmSignature == types.EventFieldTypeAddress:
		return types.SQLColumnTypeVarchar, 40, nil
		// solidity bool => sql bool
	case evmSignature == types.EventFieldTypeBool:
		return types.SQLColumnTypeBool, 0, nil
		// solidity bytes => sql bytes
		// bytesToString == true means there is a string in there so => sql varchar
	case strings.HasPrefix(evmSignature, types.EventFieldTypeBytes):
		if bytesToString {
			return types.SQLColumnTypeVarchar, 40, nil
		} else {
			return types.SQLColumnTypeByteA, 0, nil
		}
		// solidity string => sql text
	case evmSignature == types.EventFieldTypeString:
		return types.SQLColumnTypeText, 0, nil
		// solidity int or int256 => sql bigint
		// solidity int <= 32 => sql int
		// solidity int > 32 => sql numeric
	case strings.HasPrefix(evmSignature, types.EventFieldTypeInt):
		if typeSize == 0 || typeSize == 256 {
			return types.SQLColumnTypeBigInt, 0, nil
		}
		if typeSize <= 32 {
			return types.SQLColumnTypeInt, 0, nil
		} else {
			return types.SQLColumnTypeNumeric, 0, nil
		}
		// solidity uint or uint256 => sql bigint
		// solidity uint <= 16 => sql int
		// solidity uint > 16 => sql numeric
	case strings.HasPrefix(evmSignature, types.EventFieldTypeUInt):
		if typeSize == 0 || typeSize == 256 {
			return types.SQLColumnTypeBigInt, 0, nil
		}
		if typeSize <= 16 {
			return types.SQLColumnTypeInt, 0, nil
		} else {
			return types.SQLColumnTypeNumeric, 0, nil
		}
	default:
		return -1, 0, fmt.Errorf("Don't know how to map evmSignature: %s ", evmSignature)
	}
}

// getGlobalColumns returns global columns for event table structures,
// these columns will be part of every SQL event table to relate data with source events
func getGlobalFieldMappings() []*types.EventFieldMapping {
	return []*types.EventFieldMapping{
		{
			ColumnName: types.SQLColumnLabelHeight,
			Field:      types.BlockHeightLabel,
			Type:       types.EventFieldTypeString,
		},
		{
			ColumnName: types.SQLColumnLabelTxHash,
			Field:      types.TxTxHashLabel,
			Type:       types.EventFieldTypeString,
		},
		{
			ColumnName: types.SQLColumnLabelEventType,
			Field:      types.EventTypeLabel,
			Type:       types.EventFieldTypeString,
		},
		{
			ColumnName: types.SQLColumnLabelEventName,
			Field:      types.EventNameLabel,
			Type:       types.EventFieldTypeString,
		},
	}
}

// Merges tables a and b provided the intersection of their columns (by name) are identical
func mergeTables(tables ...*types.SQLTable) (*types.SQLTable, error) {
	table := &types.SQLTable{
		NotifyChannels: make(map[string][]string),
	}

	columns := make(map[string]*types.SQLTableColumn)
	notifications := make(map[string]map[string]struct{})

	for _, t := range tables {
		if t != nil {
			table.Name = t.Name
			for _, columnB := range t.Columns {
				if columnA, ok := columns[columnB.Name]; ok {
					if !columnA.Equals(columnB) {
						return nil, fmt.Errorf("cannot merge event class tables for %s because of "+
							"conflicting columns: %v and %v", t.Name, columnB, columnB)
					}
					// Just keep existing column from A - they match
				} else {
					// Add as new column
					table.Columns = append(table.Columns, columnB)
					columns[columnB.Name] = columnB
				}
			}
			for channel, columnNames := range t.NotifyChannels {
				for _, columnName := range columnNames {
					if notifications[channel] == nil {
						notifications[channel] = make(map[string]struct{})
					}
					notifications[channel][columnName] = struct{}{}
				}
			}
		}
	}

	// Merge notification channels requested by specs
	for channel, colMap := range notifications {
		for columnName := range colMap {
			table.NotifyChannels[channel] = append(table.NotifyChannels[channel], columnName)
		}
		sort.Strings(table.NotifyChannels[channel])
	}

	return table, nil
}
