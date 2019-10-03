package core

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/consensus/abci"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/process"
	"github.com/hyperledger/burrow/project"
	"github.com/hyperledger/burrow/proxy"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/lib/server"
	"github.com/hyperledger/burrow/rpc/metrics"
	"github.com/hyperledger/burrow/rpc/rpcdump"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"github.com/hyperledger/burrow/rpc/rpcinfo"
	"github.com/hyperledger/burrow/rpc/rpcquery"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"github.com/hyperledger/burrow/rpc/web3"
	"github.com/hyperledger/burrow/txs"
	"github.com/tendermint/tendermint/version"
	hex "github.com/tmthrgd/go-hex"
)

const (
	ProfilingProcessName   = "Profiling"
	DatabaseProcessName    = "Database"
	NoConsensusProcessName = "NoConsensusExecution"
	TendermintProcessName  = "Tendermint"
	StartupProcessName     = "StartupAnnouncer"
	Web3ProcessName        = "rpcConfig/web3"
	InfoProcessName        = "rpcConfig/info"
	GRPCProcessName        = "rpcConfig/GRPC"
	InternalProxyName      = "rpcConfig/Proxy"
	MetricsProcessName     = "rpcConfig/metrics"
)

func DefaultProcessLaunchers(kern *Kernel, rpcConfig *rpc.RPCConfig, proxyConfig *proxy.ProxyConfig) []process.Launcher {
	// Run announcer after Tendermint so it can get some details
	return []process.Launcher{
		ProfileLauncher(kern, rpcConfig.Profiler),
		DatabaseLauncher(kern),
		NoConsensusLauncher(kern),
		TendermintLauncher(kern),
		StartupLauncher(kern),
		Web3Launcher(kern, rpcConfig.Web3),
		InfoLauncher(kern, rpcConfig.Info),
		MetricsLauncher(kern, rpcConfig.Metrics),
		GRPCLauncher(kern, rpcConfig.GRPC),
		InternalProxyLauncher(kern, proxyConfig),
	}
}

func ProfileLauncher(kern *Kernel, conf *rpc.ServerConfig) process.Launcher {
	return process.Launcher{
		Name:    ProfilingProcessName,
		Enabled: conf.Enabled,
		Launch: func() (process.Process, error) {
			debugServer := &http.Server{
				Addr: conf.ListenAddress(),
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

func DatabaseLauncher(kern *Kernel) process.Launcher {
	return process.Launcher{
		Name:    DatabaseProcessName,
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

// Run a single uncoordinated local state
func NoConsensusLauncher(kern *Kernel) process.Launcher {
	return process.Launcher{
		Name:    NoConsensusProcessName,
		Enabled: kern.Node == nil,
		Launch: func() (process.Process, error) {
			accountState := kern.State
			nameRegState := kern.State
			nodeRegState := kern.State
			validatorSet := kern.State
			kern.Service = rpc.NewService(accountState, nameRegState, nodeRegState, kern.Blockchain, validatorSet, nil, kern.Logger)
			// TimeoutFactor scales in units of seconds
			blockDuration := time.Duration(kern.timeoutFactor * float64(time.Second))
			//proc := abci.NewProcess(kern.checker, kern.committer, kern.Blockchain, kern.txCodec, blockDuration, kern.Panic)
			proc := abci.NewProcess(kern.committer, kern.Blockchain, kern.txCodec, blockDuration, kern.Panic)
			// Provide execution accounts against backend state since we will commit immediately
			accounts := execution.NewAccounts(kern.committer, kern.keyStore, AccountsRingMutexCount)
			// Elide consensus and use a CheckTx function that immediately commits any valid transaction
			kern.Transactor = execution.NewTransactor(kern.Blockchain, kern.Emitter, accounts, proc.CheckTx, kern.txCodec,
				kern.Logger)
			return proc, nil
		},
	}
}

func TendermintLauncher(kern *Kernel) process.Launcher {
	return process.Launcher{
		Name:    TendermintProcessName,
		Enabled: kern.Node != nil,
		Launch: func() (process.Process, error) {
			const errHeader = "TendermintLauncher():"
			nodeView, err := kern.GetNodeView()
			if err != nil {
				return nil, fmt.Errorf("%s cannot get NodeView %v", errHeader, err)
			}

			kern.Blockchain.SetBlockStore(bcm.NewBlockStore(nodeView.BlockStore()))
			// Provide execution accounts against checker state so that we can assign sequence numbers
			accounts := execution.NewAccounts(kern.checker, kern.keyStore, AccountsRingMutexCount)
			// Pass transactions to Tendermint's CheckTx function for broadcast and consensus
			checkTx := kern.Node.Mempool().CheckTx
			kern.Transactor = execution.NewTransactor(kern.Blockchain, kern.Emitter, accounts, checkTx, kern.txCodec,
				kern.Logger)

			accountState := kern.State
			eventsState := kern.State
			nameRegState := kern.State
			nodeRegState := kern.State
			validatorState := kern.State
			kern.Service = rpc.NewService(accountState, nameRegState, nodeRegState, kern.Blockchain, validatorState, nodeView, kern.Logger)
			kern.EthService = rpc.NewEthService(accountState, eventsState, kern.Blockchain, validatorState, nodeView, kern.Transactor, kern.keyStore, kern.Logger)

			if err := kern.Node.Start(); err != nil {
				return nil, fmt.Errorf("%s error starting Tendermint node: %v", errHeader, err)
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
			}), nil
		},
	}
}

func StartupLauncher(kern *Kernel) process.Launcher {
	return process.Launcher{
		Name:    StartupProcessName,
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
			netAddress, err := info.NetAddress()
			if err != nil {
				return nil, err
			}
			logger := kern.Logger.With(
				"launch_time", start,
				"burrow_version", project.FullVersion(),
				"tendermint_version", version.Version,
				"validator_address", nodeView.ValidatorAddress(),
				"node_id", string(info.ID()),
				"net_address", netAddress.String(),
				"genesis_app_hash", genesisDoc.AppHash.String(),
				"genesis_hash", hex.EncodeUpperToString(genesisDoc.Hash()),
			)

			err = logger.InfoMsg("Burrow is launching. We have marmot-off.", "announce", "startup")
			return shutdown, err
		},
	}
}

func InfoLauncher(kern *Kernel, conf *rpc.ServerConfig) process.Launcher {
	return process.Launcher{
		Name:    InfoProcessName,
		Enabled: conf.Enabled,
		Launch: func() (process.Process, error) {
			listener, err := process.ListenerFromAddress(conf.ListenAddress())
			if err != nil {
				return nil, err
			}
			err = kern.registerListener(InfoProcessName, listener)
			if err != nil {
				return nil, err
			}
			server, err := rpcinfo.StartServer(kern.Service, "/websocket", listener, kern.Logger)
			if err != nil {
				return nil, err
			}
			return server, nil
		},
	}
}

func Web3Launcher(kern *Kernel, conf *rpc.ServerConfig) process.Launcher {
	return process.Launcher{
		Name:    Web3ProcessName,
		Enabled: conf.Enabled,
		Launch: func() (process.Process, error) {
			listener, err := process.ListenerFromAddress(fmt.Sprintf("%s:%s", conf.ListenHost, conf.ListenPort))
			if err != nil {
				return nil, err
			}
			err = kern.registerListener(Web3ProcessName, listener)
			if err != nil {
				return nil, err
			}

			srv, err := server.StartHTTPServer(listener, web3.NewServer(kern.EthService), kern.Logger)
			if err != nil {
				return nil, err
			}

			return srv, nil
		},
	}
}

func MetricsLauncher(kern *Kernel, conf *rpc.MetricsConfig) process.Launcher {
	return process.Launcher{
		Name:    MetricsProcessName,
		Enabled: conf.Enabled,
		Launch: func() (process.Process, error) {
			listener, err := process.ListenerFromAddress(conf.ListenAddress())
			if err != nil {
				return nil, err
			}
			err = kern.registerListener(MetricsProcessName, listener)
			if err != nil {
				return nil, err
			}
			server, err := metrics.StartServer(kern.Service, conf.MetricsPath, listener, conf.BlockSampleSize,
				kern.Logger)
			if err != nil {
				return nil, err
			}
			return server, nil
		},
	}
}

func GRPCLauncher(kern *Kernel, conf *rpc.ServerConfig) process.Launcher {
	return process.Launcher{
		Name:    GRPCProcessName,
		Enabled: conf.Enabled,
		Launch: func() (process.Process, error) {
			nodeView, err := kern.GetNodeView()
			if err != nil {
				return nil, err
			}

			listener, err := process.ListenerFromAddress(conf.ListenAddress())
			if err != nil {
				return nil, err
			}
			err = kern.registerListener(GRPCProcessName, listener)
			if err != nil {
				return nil, err
			}

			grpcServer := rpc.NewGRPCServer(kern.Logger)
			grpcServer.GetServiceInfo()

			nameRegState := kern.State
			nodeRegState := kern.State
			proposalRegState := kern.State
			rpcquery.RegisterQueryServer(grpcServer, rpcquery.NewQueryServer(kern.State, nameRegState, nodeRegState, proposalRegState,
				kern.Blockchain, kern.State, nodeView, kern.Logger))

			txCodec := txs.NewProtobufCodec()
			rpctransact.RegisterTransactServer(grpcServer,
				rpctransact.NewTransactServer(kern.State, kern.Blockchain, kern.Transactor, txCodec, kern.Logger))

			rpcevents.RegisterExecutionEventsServer(grpcServer, rpcevents.NewExecutionEventsServer(kern.State,
				kern.Emitter, kern.Blockchain, kern.Logger))

			rpcdump.RegisterDumpServer(grpcServer, rpcdump.NewDumpServer(kern.State, kern.Blockchain, kern.Logger))

			// Provides metadata about services registered
			// reflection.Register(grpcServer)

			go grpcServer.Serve(listener)

			return process.ShutdownFunc(func(ctx context.Context) error {
				grpcServer.Stop()
				// listener is closed for us
				return nil
			}), nil
		},
	}
}

func InternalProxyLauncher(kern *Kernel, conf *proxy.ProxyConfig) process.Launcher {
	return process.Launcher{
		Name:    InternalProxyName,
		Enabled: conf.Enabled,
		Launch: func() (process.Process, error) {
			nodeView, err := kern.GetNodeView()
			if err != nil {
				return nil, err
			}

			listener, err := process.ListenerFromAddress(fmt.Sprintf("%s:%s", conf.ListenHost, conf.ListenPort))
			if err != nil {
				return nil, err
			}
			err = kern.registerListener(InternalProxyName, listener)
			if err != nil {
				return nil, err
			}

			grpcServer := rpc.NewGRPCServer(kern.Logger)
			grpcServer.GetServiceInfo()

			nameRegState := kern.State
			proposalRegState := kern.State
			nodeRegState := kern.State
			rpcquery.RegisterQueryServer(grpcServer, rpcquery.NewQueryServer(kern.State, nameRegState, nodeRegState,
				proposalRegState, kern.Blockchain, kern.State, nodeView, kern.Logger))

			txCodec := txs.NewProtobufCodec()
			rpctransact.RegisterTransactServer(grpcServer,
				rpctransact.NewTransactServer(kern.State, kern.Blockchain, kern.Transactor, txCodec, kern.Logger))

			rpcevents.RegisterExecutionEventsServer(grpcServer, rpcevents.NewExecutionEventsServer(kern.State,
				kern.Emitter, kern.Blockchain, kern.Logger))

			// FIXME: start keys service

			// Provides metadata about services registered
			// reflection.Register(grpcServer)

			go grpcServer.Serve(listener)

			return process.ShutdownFunc(func(ctx context.Context) error {
				grpcServer.Stop()
				// listener is closed for us
				return nil
			}), nil
		},
	}
}
