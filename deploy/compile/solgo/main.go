package main

import (
	"fmt"
	"os"
	"path"

	"github.com/hyperledger/burrow/deploy/compile"
)

func main() {
	for _, solfile := range os.Args[1:] {
		resp, err := compile.Compile(solfile, false, nil)

		if err != nil {
			fmt.Printf("failed compile solidity: %v\n", err)
			os.Exit(1)
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
		f.WriteString("import \"github.com/tmthrgd/go-hex\"\n\n")

		for _, c := range resp.Objects {
			f.WriteString(fmt.Sprintf("var Bytecode_%s = hex.MustDecodeString(\"%s\")\n",
				c.Objectname, c.Binary.Evm.Bytecode.Object))
			f.WriteString(fmt.Sprintf("var Abi_%s = []byte(`%s`)\n",
				c.Objectname, c.Binary.Abi))
		}
	}
}
