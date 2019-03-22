package core

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/process"
	"github.com/hyperledger/burrow/project"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/metrics"
	"github.com/hyperledger/burrow/rpc/rpcdump"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"github.com/hyperledger/burrow/rpc/rpcinfo"
	"github.com/hyperledger/burrow/rpc/rpcquery"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/tendermint/version"
	hex "github.com/tmthrgd/go-hex"
)

func DefaultServices(kern *Kernel, rpc *rpc.RPCConfig, keys *keys.KeysConfig) []process.Launcher {
	// Run announcer after Tendermint so it can get some details
	return []process.Launcher{ProfileServer(kern, rpc), DatabaseServer(kern), TendermintServer(kern),
		StartupServer(kern), InfoServer(kern, rpc), MetricsServer(kern, rpc), GRPCServer(kern, rpc, keys)}
}

func ProfileServer(kern *Kernel, RPC *rpc.RPCConfig) process.Launcher {
	return process.Launcher{
		Name:    "Profiling",
		Enabled: RPC.Profiler.Enabled,
		Launch: func() (process.Process, error) {
			debugServer := &http.Server{
				Addr: ":6060",
			}
			go func() {
				err := debugServer.ListenAndServe()
				if err != nil {
					kern.Logger.InfoMsg("Error from pprof debug server", structure.ErrorKey, err)
				}
			}()
			return debugServer, nil
		},
	}
}

func DatabaseServer(kern *Kernel) process.Launcher {
	return process.Launcher{
		Name:    "Database",
		Enabled: true,
		Launch: func() (process.Process, error) {
			// Just close database
			return process.ShutdownFunc(func(ctx context.Context) error {
				kern.database.Close()
				return nil
			}), nil
		},
	}
}

func TendermintServer(kern *Kernel) process.Launcher {
	return process.Launcher{
		Name:    "Tendermint",
		Enabled: true,
		Launch: func() (process.Process, error) {
			if kern.Node == nil {
				return process.ShutdownFunc(func(ctx context.Context) error {
					kern.Logger.InfoMsg("Tendermint not enabled")
					return nil
				}), nil
			}

			if err := kern.Node.Start(); err != nil {
				return nil, fmt.Errorf("error starting Tendermint node: %v", err)
			}

			return process.ShutdownFunc(func(ctx context.Context) error {
				err := kern.Node.Stop()
				// Close tendermint database connections using our wrapper
				defer kern.Node.Close()
				if err != nil {
					return err
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-kern.Node.Quit():
					kern.Logger.InfoMsg("Tendermint Node has quit, closing DB connections...")
					return nil
				}
				return err
			}), nil
		},
	}
}

func StartupServer(kern *Kernel) process.Launcher {
	return process.Launcher{
		Name:    "Startup Announcer",
		Enabled: true,
		Launch: func() (process.Process, error) {
			start := time.Now()
			shutdown := process.ShutdownFunc(func(ctx context.Context) error {
				stop := time.Now()
				return kern.Logger.InfoMsg("Burrow is shutting down. Prepare for re-entrancy.",
					"announce", "shutdown",
					"shutdown_time", stop,
					"elapsed_run_time", stop.Sub(start).String())
			})

			if kern.Node == nil {
				return shutdown, nil
			}

			nodeView, err := kern.GetNodeView()
			if err != nil {
				return nil, err
			}

			genesisDoc := kern.Blockchain.GenesisDoc()
			info := kern.Node.NodeInfo()
			logger := kern.Logger.With(
				"launch_time", start,
				"burrow_version", project.FullVersion(),
				"tendermint_version", version.TMCoreSemVer,
				"validator_address", nodeView.ValidatorAddress(),
				"node_id", string(info.ID()),
				"net_address", info.NetAddress().String(),
				"genesis_app_hash", genesisDoc.AppHash.String(),
				"genesis_hash", hex.EncodeUpperToString(genesisDoc.Hash()),
			)

			err = logger.InfoMsg("Burrow is launching. We have marmot-off.", "announce", "startup")
			return shutdown, err
		},
	}
}

func InfoServer(kern *Kernel, RPC *rpc.RPCConfig) process.Launcher {
	return process.Launcher{
		Name:    "RPC/info",
		Enabled: RPC.Info.Enabled,
		Launch: func() (process.Process, error) {
			server, err := rpcinfo.StartServer(kern.Service, "/websocket", RPC.Info.ListenAddress, kern.Logger)
			if err != nil {
				return nil, err
			}
			return server, nil
		},
	}
}

func MetricsServer(kern *Kernel, RPC *rpc.RPCConfig) process.Launcher {
	return process.Launcher{
		Name:    "RPC/metrics",
		Enabled: RPC.Metrics.Enabled,
		Launch: func() (process.Process, error) {
			server, err := metrics.StartServer(kern.Service, RPC.Metrics.MetricsPath,
				RPC.Metrics.ListenAddress, RPC.Metrics.BlockSampleSize, kern.Logger)
			if err != nil {
				return nil, err
			}
			return server, nil
		},
	}
}

func GRPCServer(kern *Kernel, RPC *rpc.RPCConfig, keyConfig *keys.KeysConfig) process.Launcher {
	return process.Launcher{
		Name:    "RPC/GRPC",
		Enabled: RPC.GRPC.Enabled,
		Launch: func() (process.Process, error) {
			nodeView, err := kern.GetNodeView()
			if err != nil {
				return nil, err
			}

			listen, err := net.Listen("tcp", RPC.GRPC.ListenAddress)
			if err != nil {
				return nil, err
			}

			grpcServer := rpc.NewGRPCServer(kern.Logger)
			var ks *keys.KeyStore
			if kern.keyStore != nil {
				ks = kern.keyStore
			}

			if keyConfig.GRPCServiceEnabled {
				if kern.keyStore == nil {
					ks = keys.NewKeyStore(keyConfig.KeysDirectory, keyConfig.AllowBadFilePermissions)
				}
				keys.RegisterKeysServer(grpcServer, ks)
			}

			nameRegState := kern.State
			proposalRegState := kern.State
			rpcquery.RegisterQueryServer(grpcServer, rpcquery.NewQueryServer(kern.State, nameRegState, proposalRegState,
				kern.Blockchain, kern.State, nodeView, kern.Logger))

			txCodec := txs.NewAminoCodec()
			rpctransact.RegisterTransactServer(grpcServer, rpctransact.NewTransactServer(kern.Transactor, txCodec))

			rpcevents.RegisterExecutionEventsServer(grpcServer, rpcevents.NewExecutionEventsServer(kern.State,
				kern.Emitter, kern.Blockchain, kern.Logger))

			rpcdump.RegisterDumpServer(grpcServer, rpcdump.NewDumpServer(kern.State, kern.Blockchain, kern.Logger))

			// Provides metadata about services registered
			// reflection.Register(grpcServer)

			go grpcServer.Serve(listen)

			return process.ShutdownFunc(func(ctx context.Context) error {
				grpcServer.Stop()
				// listener is closed for us
				return nil
			}), nil
		},
	}
}
