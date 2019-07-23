package dump

import (
	"encoding/json"
	"io"
)

type NullSink struct{}

func (NullSink) Send(*Dump) error {
	return nil
}

type CollectSink struct {
	Rows    []string
	Current int
}

func (c *CollectSink) Send(d *Dump) error {
	bs, _ := json.Marshal(d)

	c.Rows = append(c.Rows, string(bs))

	return nil
}

func (c *CollectSink) Recv() (d *Dump, err error) {
	if c.Current >= len(c.Rows) {
		c.Current = 0
		return nil, io.EOF
	}
	d = new(Dump)
	err = json.Unmarshal([]byte(c.Rows[c.Current]), d)
	if err == nil {
		c.Current++
	}
	return
}
