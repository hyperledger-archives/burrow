package def

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArgMap(t *testing.T) {
	mp := argMap(&CallArg{
		Address: "fooo",
	})
	fmt.Println(mp)
	assert.Equal(t, "fooo", mp["Address"])
	assert.Len(t, mp, 7)
}
