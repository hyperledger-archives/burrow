package commands

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hyperledger/burrow/deploy/compile"
	"github.com/hyperledger/burrow/execution/evm/abi"
	"github.com/hyperledger/burrow/rpc/rpcquery"
	cli "github.com/jawher/mow.cli"
	"google.golang.org/grpc"
)

// Accounts lists all the accounts in a chain, alongside with any metadata like contract name and ABI
func Accounts(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		chainURLOpt := cmd.StringOpt("c chain", "127.0.0.1:10997", "chain to be used in IP:PORT format")
		timeoutOpt := cmd.IntOpt("t timeout", 0, "Timeout in seconds")

		cmd.Action = func() {
			ctx, cancel := context.WithCancel(context.Background())
			if *timeoutOpt != 0 {
				timeout := time.Duration(*timeoutOpt) * time.Second
				ctx, cancel = context.WithTimeout(context.Background(), timeout)
			}
			defer cancel()

			var opts []grpc.DialOption
			opts = append(opts, grpc.WithInsecure())
			conn, err := grpc.DialContext(ctx, *chainURLOpt, opts...)
			if err != nil {
				output.Fatalf("failed to connect: %v", err)
			}

			qCli := rpcquery.NewQueryClient(conn)

			stream, err := qCli.ListAccounts(context.Background(), &rpcquery.ListAccountsParam{})
			if err != nil {
				output.Fatalf("failed to list accounts: %v", err)
			}

			for acc, err := stream.Recv(); err == nil; acc, err = stream.Recv() {
				output.Printf("Account: %s\n  Sequence: %d",
					acc.Address.String(), acc.Sequence)

				if len(acc.PublicKey.PublicKey) > 0 {
					output.Printf("  Public Key: %s\n", acc.PublicKey.String())
				}
				if acc.WASMCode != nil && len(acc.WASMCode) > 0 {
					output.Printf("  WASM Code Hash: %s", acc.CodeHash.String())
				}
				if acc.EVMCode != nil && len(acc.EVMCode) > 0 {
					output.Printf("  EVM Code Hash: %s", acc.CodeHash.String())
				}

				meta, err := qCli.GetMetadata(context.Background(), &rpcquery.GetMetadataParam{Address: &acc.Address})
				if err != nil {
					output.Fatalf("failed to get metadata for %s: %v", acc.Address, err)
				}
				if meta.Metadata != "" {
					var metadata compile.Metadata
					err = json.Unmarshal([]byte(meta.Metadata), &metadata)
					if err != nil {
						output.Fatalf("failed to unmarshal metadata %s: %v", meta.Metadata, err)
					}

					output.Printf("  Contract Name: %s", metadata.ContractName)
					output.Printf("  Source File: %s", metadata.SourceFile)
					output.Printf("  Compiler version: %s", metadata.CompilerVersion)

					spec, err := abi.ReadSpec(metadata.Abi)
					if err != nil {
						output.Fatalf("failed to unmarshall abi %s: %v", string(metadata.Abi), err)
					}

					if len(spec.Functions) > 0 {
						output.Printf("  Functions:")
						for _, f := range spec.Functions {
							output.Printf("    %s", f.String())
						}
					}

					if len(spec.EventsByID) > 0 {
						output.Printf("  Events:")
						for _, e := range spec.EventsByID {
							output.Printf("    %s", e.String())
						}
					}
				}

				output.Printf("")
			}
		}
	}
}
