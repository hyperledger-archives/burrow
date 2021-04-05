package commands

import (
	"fmt"
	"os"
	"path"

	"github.com/hyperledger/burrow/deploy/compile"
	"github.com/hyperledger/burrow/logging"
	cli "github.com/jawher/mow.cli"
)

// Currently this just compiles to Go fixtures - it might make sense to extend it to take a text template for output
// if it is convenient to expose our compiler wrappers outside of burrow deploy
func Compile(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		wasmOpt := cmd.BoolOpt("w wasm", false, "Use solang rather than solc")
		sourceArg := cmd.StringsArg("SOURCE", nil, "Solidity source files to compile")
		cmd.Spec = "[--wasm] SOURCE..."

		cmd.Action = func() {
			for _, solfile := range *sourceArg {
				var resp *compile.Response
				var err error

				if *wasmOpt {
					resp, err = compile.WASM(solfile, "", logging.NewNoopLogger())
					if err != nil {
						output.Fatalf("failed compile solidity to wasm: %v\n", err)
					}
				} else {
					resp, err = compile.EVM(solfile, false, "", nil, logging.NewNoopLogger())
					if err != nil {
						output.Fatalf("failed compile solidity: %v\n", err)
					}
				}

				if resp.Error != "" {
					output.Fatalf(resp.Error)
				}

				if resp.Warning != "" {
					output.Printf(resp.Warning)
				}

				f, err := os.Create(solfile + ".go")
				if err != nil {
					output.Fatalf("failed to create go file: %v\n", err)
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
	}
}
