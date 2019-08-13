package types

import (
	"fmt"
	"testing"

	"github.com/hyperledger/burrow/config/source"
)

func TestEventTablesSchema(t *testing.T) {
	schema := ProjectionSpecSchema()
	fmt.Println(source.JSONString(schema))
}
