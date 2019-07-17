package sqlsol

import (
	"fmt"

	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/vent/types"
)

// GenerateSpecFromAbis creates a simple spec which just logs all events
func GenerateSpecFromAbis(spec *abi.Spec) ([]*types.EventClass, error) {
	type field struct {
		Type   abi.EVMType
		Events []string
	}

	fields := make(map[string]field)

	for _, ev := range spec.EventsByID {
		for _, in := range ev.Inputs {
			field, ok := fields[in.Name]
			if ok {
				if field.Type != in.EVM {
					if field.Type.ImplicitCast(in.EVM) {
						field.Type = in.EVM
					} else if !in.EVM.ImplicitCast(field.Type) {
						fmt.Printf("WARNING: field %s in event %s has different definitions in events %v (%s rather than %s)\n", in.Name, ev.Name, field.Events, field.Type, in.EVM)
					}
				} else {
					field.Events = append(field.Events, ev.Name)
				}
			} else {
				field.Type = in.EVM
				field.Events = []string{ev.Name}
			}
			fields[in.Name] = field
		}
	}

	ev := types.EventClass{
		TableName:     "event",
		Filter:        "EventType = 'LogEvent'",
		FieldMappings: make([]*types.EventFieldMapping, len(fields)),
	}

	i := 0

	for name, field := range fields {
		ev.FieldMappings[i] = &types.EventFieldMapping{
			Field:      name,
			ColumnName: name,
			Type:       field.Type.GetSignature(),
			Primary:    false,
		}
		i++
	}

	return []*types.EventClass{&ev}, nil
}
