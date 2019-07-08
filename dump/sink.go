package dump

type NullSink struct{}

func (NullSink) Send(*Dump) error {
	return nil
}
