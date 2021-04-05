package commands

import (
	"fmt"
	"strings"

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
						for _, f := range spec.Functions {
							fmt.Printf("func %x: %s\n", f.FunctionID, f.String())
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
							args := make([]string, len(fspec.Inputs))
							intf := make([]interface{}, len(args))
							for i := range args {
								intf[i] = &args[i]
							}
							err = abi.Unpack(fspec.Inputs, bs[len(funcid):], intf...)
							if err != nil {
								output.Fatalf("unable to decode function %s: %v\n", name, err)
							}
							// prepend function argument names
							for i, a := range args {
								if fspec.Inputs[i].Name != "" {
									args[i] = fspec.Inputs[i].Name + "=" + a
								}
							}

							output.Printf(fmt.Sprintf("%s(%s)", name, strings.Join(args, ",")))
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

					args := make([]string, len(fspec.Outputs))
					intf := make([]interface{}, len(args))
					for i := range args {
						intf[i] = &args[i]
					}
					err = abi.Unpack(fspec.Outputs, bs, intf...)
					if err != nil {
						output.Fatalf("unable to decode function %s: %v\n", *fname, err)
					}
					// prepend function return value names
					for i, a := range args {
						if fspec.Outputs[i].Name != "" {
							args[i] = fspec.Outputs[i].Name + "=" + a
						}
					}

					output.Printf(strings.Join(args, "\n"))
				}
			})
	}
}
