package dump

import (
	"encoding/json"
	"io"
	"os"
)

type Reader interface {
	Next() (*Dump, error)
}

type FileReader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewFileReader(filename string) (Reader, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &FileReader{file: f}, nil
}

func (f *FileReader) Next() (*Dump, error) {
	var row Dump
	var err error

	if f.decoder != nil {
		err = f.decoder.Decode(&row)
	} else {
		_, err = cdc.UnmarshalBinaryLengthPrefixedReader(f.file, &row, 0)

		if err != nil && err != io.EOF && f.decoder == nil {
			f.file.Seek(0, 0)

			f.decoder = json.NewDecoder(f.file)

			return f.Next()
		}
	}

	if err == io.EOF {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &row, err
}
