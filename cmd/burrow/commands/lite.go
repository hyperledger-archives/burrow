package commands

import (
	"context"
	"time"

	"github.com/hyperledger/burrow/rpc/lite"
	cli "github.com/jawher/mow.cli"
	tmlite "github.com/tendermint/tendermint/lite2"
	"github.com/tendermint/tendermint/lite2/provider"
	dbs "github.com/tendermint/tendermint/lite2/store/db"
	dbm "github.com/tendermint/tm-db"
	"google.golang.org/grpc"
)

// Lite launches a burrow light client against a single primary
func Lite(output Output) func(cmd *cli.Cmd) {
	return func(cmd *cli.Cmd) {
		chainIDOpt := cmd.StringOpt("id", "", "chain ID")
		providerURLOpt := cmd.StringOpt("p provider", "127.0.0.1:10997", "chain to be used in IP:PORT format")
		witnessURLOpt := cmd.StringsOpt("w witness", []string{}, "to verify against primary node")
		timeoutOpt := cmd.IntOpt("t timeout", 0, "Timeout in seconds")

		cmd.Action = func() {
			ctx, cancel := context.WithCancel(context.Background())
			if *timeoutOpt != 0 {
				timeout := time.Duration(*timeoutOpt) * time.Second
				ctx, cancel = context.WithTimeout(context.Background(), timeout)
			}
			defer cancel()

			chainID := *chainIDOpt
			prov, err := newProvider(ctx, *providerURLOpt, chainID)
			if err != nil {
				output.Fatalf("failed to connect to provider: %v", err)
			}

			var witnesses []provider.Provider
			for _, url := range *witnessURLOpt {
				w, err := newProvider(ctx, url, chainID)
				if err != nil {
					output.Fatalf("failed to connect to witness: %v", err)
				}
				witnesses = append(witnesses, w)
			}

			db := dbm.NewMemDB()
			client, err := tmlite.NewClient(
				chainID,
				tmlite.TrustOptions{
					Period: time.Hour,
					Height: 1,
					Hash:   make([]byte, 32),
				},
				prov,
				witnesses,
				dbs.New(db, chainID),
			)
			if err != nil {
				output.Fatalf("error running light client: %v", err)
			}

			client.Start()
		}
	}
}

func newProvider(ctx context.Context, url, id string) (*lite.Provider, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.DialContext(ctx, url, opts...)
	if err != nil {
		return nil, err
	}
	return lite.NewProvider(conn, id), nil
}
