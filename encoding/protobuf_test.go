package encoding

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteMessages(t *testing.T) {
	var fErr error

	f := func(msgs []TestMessage) bool {
		var n, written, read int
		buf := new(bytes.Buffer)
		// encode
		for _, msg := range msgs {
			n, fErr = WriteMessage(buf, &msg)
			written += n
			if fErr != nil {
				return false
			}
		}

		// Require non nil for equality check later
		msgOut := []TestMessage{}
		// decode
		for {
			msg := new(TestMessage)
			n, fErr = ReadMessage(buf, msg)
			read += n
			if fErr != nil {
				if fErr == io.EOF {
					fErr = nil
					break
				}
				return false
			}

			msgOut = append(msgOut, *msg)
		}

		return assert.Equal(t, msgs, msgOut, "messages read should equal those written") &&
			assert.Equal(t, written, read, "should read the same number of bytes as written")
	}
	err := quick.Check(f, &quick.Config{
		// Takes about a second on my machine
		MaxCount: 9994,
		Rand:     rand.New(rand.NewSource(320492384234234)),
		// Custom value function because arbitrary values for some of the XXX fields can mess things up for proto.Marshal
		Values: func(values []reflect.Value, rand *rand.Rand) {
			for i := 0; i < len(values); i++ {
				msgs := make([]TestMessage, rand.Intn(200))
				for j := range msgs {
					msgs[j] = TestMessage{Type: rand.Uint32(), Amount: rand.Uint64()}
				}
				values[i] = reflect.ValueOf(msgs)
			}
		},
	})
	if err != nil {
		var literal string
		err := err.(*quick.CheckError)
		for _, in := range err.In {
			var str []string
			for _, v := range in.([]TestMessage) {
				str = append(str, v.String())
			}
			literal = fmt.Sprintf("msgs := []TestMessage{%s}", strings.Join(str, ", "))
		}
		t.Logf("CheckError with:\n%s", literal)
	}
	require.NoError(t, fErr)
}
