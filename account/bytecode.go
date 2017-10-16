package account

import (
	"encoding/json"

	"github.com/tmthrgd/go-hex"
)

// TODO: write a simple lexer that prints the opcodes. Each byte is either an OpCode or part of the Word256 argument
// to Push[1...32]
type Bytecode []byte

func (bc Bytecode) Bytes() []byte {
	return bc[:]
}

func (bc Bytecode) String() string {
	return hex.EncodeToString(bc[:])

}
func (bc Bytecode) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeUpperToString(bc[:]))
}

func (bc Bytecode) UnmarshalJSON(data []byte) error {
	str := new(string)
	err := json.Unmarshal(data, str)
	if err != nil {
		return err
	}
	_, err = hex.Decode(bc[:], []byte(*str))
	if err != nil {
		return err
	}
	return nil
}
