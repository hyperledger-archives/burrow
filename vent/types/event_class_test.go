package types

import (
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/config/source"
)

func TestEventTablesSchema(t *testing.T) {
	schema := EventSpecSchema()
	fmt.Println(source.JSONString(schema))
}
