package txs

import (
	"bytes"
	"sync"

	"github.com/tendermint/go-wire"
)

type goWireCodec struct {
	// Worth it? Possibly not, but we need to instantiate a codec though so...
	bufferPool sync.Pool
}

func NewGoWireCodec() *goWireCodec {
	return &goWireCodec{
		bufferPool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

func (gwc *goWireCodec) EncodeTx(tx Tx) ([]byte, error) {
	var n int
	var err error
	buf := gwc.bufferPool.Get().(*bytes.Buffer)
	defer gwc.recycle(buf)
	wire.WriteBinary(struct{ Tx }{tx}, buf, &n, &err)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// panic on err
func (gwc *goWireCodec) DecodeTx(txBytes []byte) (Tx, error) {
	var n int
	var err error
	tx := new(Tx)
	buf := bytes.NewBuffer(txBytes)
	wire.ReadBinaryPtr(tx, buf, len(txBytes), &n, &err)
	if err != nil {
		return nil, err
	}
	return *tx, nil
}

func (gwc *goWireCodec) recycle(buf *bytes.Buffer) {
	buf.Reset()
	gwc.bufferPool.Put(buf)
}
