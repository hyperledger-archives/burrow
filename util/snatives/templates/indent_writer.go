// Copyright 2019 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
