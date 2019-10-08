// Copyright Monax Industries Limited
// SPDX-License-Identifier: Apache-2.0

package templates

import "io"

const newLine = byte('\n')

type indentWriter struct {
	writer      io.Writer
	indentLevel uint
	indentBytes []byte
	indent      bool
}

var _ io.Writer = (*indentWriter)(nil)

// indentWriter indents all lines written to it with a specified indent string
// indented the specified number of indents
func NewIndentWriter(indentLevel uint, indentString string,
	writer io.Writer) *indentWriter {
	return &indentWriter{
		writer:      writer,
		indentLevel: indentLevel,
		indentBytes: []byte(indentString),
		indent:      true,
	}
}

func (iw *indentWriter) Write(p []byte) (int, error) {
	bs := make([]byte, 0, len(p))
	for _, b := range p {
		if iw.indent {
			for i := uint(0); i < iw.indentLevel; i++ {
				bs = append(bs, iw.indentBytes...)
			}
			iw.indent = false
		}
		if b == newLine {
			iw.indent = true
		}
		bs = append(bs, b)
	}
	return iw.writer.Write(bs)
}

func (iw *indentWriter) SetIndent(level uint) {
	iw.indentLevel = level
}
