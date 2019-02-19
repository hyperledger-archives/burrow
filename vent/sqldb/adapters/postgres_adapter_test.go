package adapters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgresAdapter_CreateTriggerQuery(t *testing.T) {
	assert.Equal(t, `'Address', NEW."Address", 'Name', NEW."Name", 'Index', NEW."Index"`,
		jsonBuildObjectArgs("NEW", []string{"Address", "Name", "Index"}))
}
