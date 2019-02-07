package types

import (
	"fmt"
	"github.com/hyperledger/burrow/config/source"
	"testing"
)

func TestEventTablesSchema(t *testing.T) {
	schema := EventSpecSchema()
	fmt.Println(source.JSONString(schema))
}