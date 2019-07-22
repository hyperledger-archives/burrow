package main

import (
	"flag"
	"fmt"
	"os"
	"path"

	"github.com/hyperledger/burrow/deploy/compile"
	"github.com/hyperledger/burrow/logging"
)

func main() {
	wasmPtr := flag.Bool("wasm", false, "Use solang rather than solc")
	flag.Parse()

	for _, solfile := range flag.Args() {
		var resp *compile.Response
		var err error

		if *wasmPtr {
			resp, err = compile.WASM(solfile, "", logging.NewNoopLogger())
			if err != nil {
				fmt.Printf("failed compile solidity to wasm: %v\n", err)
				os.Exit(1)
			}
		} else {
			resp, err = compile.EVM(solfile, false, "", nil, logging.NewNoopLogger())
			if err != nil {
				fmt.Printf("failed compile solidity: %v\n", err)
				os.Exit(1)
			}
		}

		if resp.Error != "" {
			fmt.Print(resp.Error)
			os.Exit(1)
		}

		if resp.Warning != "" {
			fmt.Print(resp.Warning)
			os.Exit(1)
		}

		f, err := os.Create(solfile + ".go")
		if err != nil {
			fmt.Printf("failed to create go file: %v\n", err)
			os.Exit(1)
		}

		f.WriteString(fmt.Sprintf("package %s\n\n", path.Base(path.Dir(solfile))))
		f.WriteString("import hex \"github.com/tmthrgd/go-hex\"\n\n")

		for _, c := range resp.Objects {
			code := c.Contract.Evm.Bytecode.Object
			if code == "" {
				code = c.Contract.EWasm.Wasm
			}
			f.WriteString(fmt.Sprintf("var Bytecode_%s = hex.MustDecodeString(\"%s\")\n",
				c.Objectname, code))
			if c.Contract.Evm.DeployedBytecode.Object != "" {
				f.WriteString(fmt.Sprintf("var DeployedBytecode_%s = hex.MustDecodeString(\"%s\")\n",
					c.Objectname, c.Contract.Evm.DeployedBytecode.Object))
			}
			f.WriteString(fmt.Sprintf("var Abi_%s = []byte(`%s`)\n",
				c.Objectname, c.Contract.Abi))
		}
	}
}
