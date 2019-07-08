package dump

import "io"

// Implements both Sink and Source
type Pipe chan msg

type msg struct {
	dump *Dump
	err  error
}

func (p Pipe) Recv() (*Dump, error) {
	msg, ok := <-p
	if !ok {
		return nil, io.EOF
	}
	if msg.err != nil {
		return nil, msg.err
	}
	return msg.dump, nil
}

func (p Pipe) Send(dump *Dump) error {
	p <- msg{dump: dump}
	return nil
}
