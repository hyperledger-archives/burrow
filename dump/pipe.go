package dump

import "io"

// Implements both Sender and Receiver
type Pipe chan msg

type msg struct {
	dump *Dump
	err  error
}

type Sender interface {
	Send(*Dump) error
}

type Receiver interface {
	Recv() (*Dump, error)
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
