package dump

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/hyperledger/burrow/encoding"
)

type Source interface {
	Recv() (*Dump, error)
}

type StreamReader struct {
	reader io.Reader
	decode func(*Dump) error
}

func NewFileReader(filename string) (Source, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	decoder, err := decoderFor(f)
	if err != nil {
		return nil, err
	}

	return NewStreamReader(f, decoder)
}

func NewProtobufReader(reader io.Reader) (*StreamReader, error) {
	return NewStreamReader(reader, protobufDecoder(reader))
}

func NewJSONReader(reader io.Reader) (*StreamReader, error) {
	return NewStreamReader(reader, jsonDecoder(reader))
}

func NewStreamReader(reader io.Reader, decode func(*Dump) error) (*StreamReader, error) {
	return &StreamReader{
		reader: reader,
		decode: decode,
	}, nil
}

func (sr *StreamReader) Recv() (*Dump, error) {
	row := new(Dump)

	err := sr.decode(row)
	if err != nil {
		return nil, err
	}

	return row, nil
}

func protobufDecoder(r io.Reader) func(*Dump) error {
	return func(row *Dump) error {
		_, err := encoding.ReadMessage(r, row)
		return err
	}
}
func jsonDecoder(r io.Reader) func(*Dump) error {
	decoder := json.NewDecoder(r)
	return func(dump *Dump) error {
		return decoder.Decode(dump)
	}
}

// Detects whether dump file appears to be protobuf or JSON encoded by trying to decode the first row with each
func decoderFor(f *os.File) (func(*Dump) error, error) {
	defer f.Seek(0, 0)

	jsonErr := json.NewDecoder(f).Decode(&Dump{})
	if jsonErr == nil || jsonErr == io.EOF {
		return jsonDecoder(f), nil
	}

	_, binErr := encoding.ReadMessage(f, &Dump{})
	if binErr != nil && binErr != io.EOF {
		return nil, fmt.Errorf("could decode first row of dump file as protobuf (%v) or JSON (%v)",
			binErr, jsonErr)
	}

	return protobufDecoder(f), nil
}
