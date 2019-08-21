package commands

import (
	"fmt"

	"github.com/hyperledger/burrow/execution/evm/abi"
	cli "github.com/jawher/mow.cli"
	hex "github.com/tmthrgd/go-hex"
)

// Abi is a command line tool for ABI encoding and decoding. Event encoding/decoding still be added
func Abi(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		cmd.Command("list", "List the functions and events",
			func(cmd *cli.Cmd) {
				dirs := cmd.StringsArg("DIR", nil, "ABI file or directory")

				cmd.Action = func() {
					for _, d := range *dirs {
						output.Printf("In %s\n", d)
						spec, err := abi.LoadPath(d)
						if err != nil {
							output.Printf("could not read %s: %v", d, err)
						}
						for id, e := range spec.EventsByID {
							fmt.Printf("event %s: %s\n", id, e.String())
						}
						for name, f := range spec.Functions {
							fmt.Printf("func %x: %s\n", f.FunctionID, f.String(name))
						}
					}
				}
			})

		cmd.Command("encode-function-call", "ABI encode function call",
			func(cmd *cli.Cmd) {
				abiPath := cmd.StringOpt("abi", ".", "ABI file or directory")
				fname := cmd.StringArg("FUNCTION", "", "Function name")
				args := cmd.StringsArg("ARGS", nil, "Function arguments")

				cmd.Spec = "--abi=<path> FUNCTION [ARGS...]"

				cmd.Action = func() {
					spec, err := abi.LoadPath(*abiPath)
					if err != nil {
						output.Fatalf("could not read %v: %v", *abiPath, err)
					}

					argsInInterface := make([]interface{}, len(*args))
					for i, a := range *args {
						argsInInterface[i] = a
					}

					data, _, err := spec.Pack(*fname, argsInInterface...)
					if err != nil {
						output.Fatalf("could not encode function call %v", err)
					}

					output.Printf("%X\n", data)
				}
			})

		cmd.Command("decode-function-call", "ABI decode function call",
			func(cmd *cli.Cmd) {
				abiPath := cmd.StringOpt("abi", ".", "ABI file or directory")
				data := cmd.StringArg("DATA", "", "Encoded function call")

				cmd.Action = func() {
					spec, err := abi.LoadPath(*abiPath)
					if err != nil {
						output.Fatalf("could not read %v: %v", *abiPath, err)
					}

					bs, err := hex.DecodeString(*data)
					if err != nil {
						output.Fatalf("could not hex decode %s: %v", data, err)
					}

					var funcid abi.FunctionID
					copy(funcid[:], bs)
					found := false
					for name, fspec := range spec.Functions {
						if fspec.FunctionID == funcid {
							args := make([]interface{}, len(fspec.Inputs))
							for i := range args {
								args[i] = new(string)
							}
							err = abi.Unpack(fspec.Inputs, bs[4:], args...)
							if err != nil {
								output.Fatalf("unable to decode function %s: %v\n", name, err)
							}
							decoded := name + "("
							for i, a := range args {
								if i > 0 {
									decoded += ","
								}
								decoded += *(a.(*string))
							}
							decoded += ")"

							output.Fatalf(decoded)
						}

					}

					if !found {
						output.Fatalf("could not find function %X\n", funcid)
					}
				}
			})

		cmd.Command("decode-function-return", "ABI decode function return",
			func(cmd *cli.Cmd) {
				abiPath := cmd.StringOpt("abi", ".", "ABI file or directory")
				fname := cmd.StringArg("FUNCTION", "", "Function name")
				data := cmd.StringArg("DATA", "", "Encoded function call")

				cmd.Action = func() {
					spec, err := abi.LoadPath(*abiPath)
					if err != nil {
						output.Fatalf("could not read %v: %v", *abiPath, err)
					}

					bs, err := hex.DecodeString(*data)
					if err != nil {
						output.Fatalf("could not hex decode %s: %v", data, err)
					}

					fspec, ok := spec.Functions[*fname]
					if !ok {
						output.Fatalf("no such function %s\n", *fname)
					}

					args := make([]interface{}, len(fspec.Outputs))
					for i := range args {
						args[i] = new(string)
					}
					err = abi.Unpack(fspec.Outputs, bs, args...)
					if err != nil {
						output.Fatalf("unable to decode function %s: %v\n", *fname, err)
					}
					decoded := "("
					for i, a := range args {
						if i > 0 {
							decoded += ","
						}
						decoded += *(a.(*string))
					}
					decoded += ")"

					output.Fatalf(decoded)
				}
			})
	}
}
