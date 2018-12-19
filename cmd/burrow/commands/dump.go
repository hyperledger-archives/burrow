package commands

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/hyperledger/burrow/rpc/rpcdump"
	amino "github.com/tendermint/go-amino"

	cli "github.com/jawher/mow.cli"
	"google.golang.org/grpc"
)

var cdc = amino.NewCodec()

func Dump(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		chainUrlOpt := cmd.StringOpt("u chain-url", "127.0.0.1:10997", "chain-url to be used in IP:PORT format")
		heightOpt := cmd.IntOpt("h height", 0, "Block height to dump to, defaults to latest block height")
		filename := cmd.StringArg("FILE", "", "Save dump here")

		cmd.Action = func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			var opts []grpc.DialOption
			opts = append(opts, grpc.WithInsecure())
			conn, err := grpc.Dial(*chainUrlOpt, opts...)
			if err != nil {
				output.Fatalf("failed to connect: %v", err)
				return
			}
			dc := rpcdump.NewDumpClient(conn)

			dump, err := dc.GetDump(ctx, &rpcdump.GetDumpParam{Height: uint64(*heightOpt)})
			if err != nil {
				output.Fatalf("failed to retrieve dump: %v", err)
				return
			}

			f, err := os.OpenFile(*filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				output.Fatalf("%s: failed to save dump: %v", *filename, err)
				return
			}

			for {
				resp, err := dump.Recv()
				if err == io.EOF {
					break
				}
				bs, err := cdc.MarshalBinaryLengthPrefixed(resp)
				if err != nil {
					output.Fatalf("failed to marshall dump: %v", *filename, err)
				}

				n, err := f.Write(bs)
				if err == nil && n < len(bs) {
					output.Fatalf("%s: failed to save dump: %v", *filename, err)
				}
			}

			if err := f.Close(); err != nil {
				output.Fatalf("%s: failed to save dump: %v", *filename, err)
			}
		}
	}
}
