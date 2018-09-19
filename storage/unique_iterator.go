package storage

import "bytes"

type uniqueIterator struct {
	source  KVIterator
	prevKey []byte
}

func Uniq(source KVIterator) *uniqueIterator {
	return &uniqueIterator{
		source: source,
	}
}

func (ui *uniqueIterator) Domain() ([]byte, []byte) {
	return ui.source.Domain()
}

func (ui *uniqueIterator) Valid() bool {
	return ui.source.Valid()
}

func (ui *uniqueIterator) Next() {
	ui.prevKey = ui.Key()
	ui.source.Next()
	// Skip elements with the same key a previous
	for ui.source.Valid() && bytes.Equal(ui.Key(), ui.prevKey) {
		ui.source.Next()
	}
}

func (ui *uniqueIterator) Key() []byte {
	return ui.source.Key()
}

func (ui *uniqueIterator) Value() []byte {
	return ui.source.Value()
}

func (ui *uniqueIterator) Close() {
	ui.source.Close()
}
