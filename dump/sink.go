package dump

import (
	"encoding/json"
)

type NullSink struct{}

func (NullSink) Send(*Dump) error {
	return nil
}

type CollectSink struct {
	Rows []string
}

func (c *CollectSink) Send(d *Dump) error {
	bs, _ := json.Marshal(d)

	c.Rows = append(c.Rows, string(bs))

	return nil
}
