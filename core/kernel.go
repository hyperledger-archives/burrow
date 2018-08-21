// Copyright 2017 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/consensus/tendermint/abci"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/process"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/rpc/metrics"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"github.com/hyperledger/burrow/rpc/rpcinfo"
	"github.com/hyperledger/burrow/rpc/rpcquery"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"github.com/hyperledger/burrow/txs"
	"github.com/streadway/simpleuuid"
	tmConfig "github.com/tendermint/tendermint/config"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/node"
	tmTypes "github.com/tendermint/tendermint/types"
)

const (
	FirstBlockTimeout      = 3 * time.Second
	CooldownTime           = 1000 * time.Millisecond
	ServerShutdownTimeout  = 1000 * time.Millisecond
	LoggingCallerDepth     = 5
	AccountsRingMutexCount = 100
)

// Kernel is the root structure of Burrow
type Kernel struct {
	// Expose these public-facing interfaces to allow programmatic extension of the Kernel by other projects
	Emitter    event.Emitter
	Service    *rpc.Service
	Launchers  []process.Launcher
	State      *execution.State
	Blockchain *bcm.Blockchain
	Node       *tendermint.Node
	// Time-based UUID randomly generated each time Burrow is started
	RunID          simpleuuid.UUID
	Logger         *logging.Logger
	nodeInfo       string
	processes      map[string]process.Process
	shutdownNotify chan struct{}
	shutdownOnce   sync.Once
}

func NewKernel(ctx context.Context, keyClient keys.KeyClient, privValidator tmTypes.PrivValidator,
	genesisDoc *genesis.GenesisDoc, tmConf *tmConfig.Config, rpcConfig *rpc.RPCConfig, keyConfig *keys.KeysConfig,
	keyStore *keys.KeyStore, exeOptions []execution.ExecutionOption, logger *logging.Logger) (*Kernel, error) {

	var err error
	kern := &Kernel{
		processes:      make(map[string]process.Process),
		shutdownNotify: make(chan struct{}),
	}
	// Create a random ID based on start time
	kern.RunID, err = simpleuuid.NewTime(time.Now())

	logger = logger.WithScope("NewKernel()").With(structure.TimeKey, log.DefaultTimestampUTC,
		structure.RunId, kern.RunID.String())
	tmLogger := logger.With(structure.CallerKey, log.Caller(LoggingCallerDepth+1))
	kern.Logger = logger.WithInfo(structure.CallerKey, log.Caller(LoggingCallerDepth))
	stateDB := dbm.NewDB("burrow_state", dbm.GoLevelDBBackend, tmConf.DBDir())

	kern.Blockchain, err = bcm.LoadOrNewBlockchain(stateDB, genesisDoc, kern.Logger)
	if err != nil {
		return nil, fmt.Errorf("error creating or loading blockchain state: %v", err)
	}

	// These should be in sync unless we are at the genesis block
	if kern.Blockchain.LastBlockHeight() > 0 {
		kern.Logger.InfoMsg("Loading application state")
		kern.State, err = execution.LoadState(stateDB, kern.Blockchain.AppHashAfterLastBlock())
		if err != nil {
			return nil, fmt.Errorf("could not load persisted execution state at hash 0x%X: %v",
				kern.Blockchain.AppHashAfterLastBlock(), err)
		}
	} else {
		kern.State, err = execution.MakeGenesisState(stateDB, genesisDoc)
	}
	kern.Logger.InfoMsg("State loading successful")

	txCodec := txs.NewAminoCodec()
	tmGenesisDoc := tendermint.DeriveGenesisDoc(genesisDoc)
	checker := execution.NewBatchChecker(kern.State, kern.Blockchain, kern.Logger)

	kern.Emitter = event.NewEmitter(kern.Logger)
	committer := execution.NewBatchCommitter(kern.State, kern.Blockchain, kern.Emitter, kern.Logger, exeOptions...)

	kern.nodeInfo = fmt.Sprintf("Burrow_%s_ValidatorID:%X", genesisDoc.ChainID(), privValidator.GetAddress())
	app := abci.NewApp(kern.nodeInfo, kern.Blockchain, checker, committer, txCodec, kern.Panic, logger)
	// We could use this to provide/register our own metrics (though this will register them with us). Unfortunately
	// Tendermint currently ignores the metrics passed unless its own server is turned on.
	metricsProvider := node.DefaultMetricsProvider(&tmConfig.InstrumentationConfig{
		Prometheus:           false,
		PrometheusListenAddr: "",
	})
	kern.Node, err = tendermint.NewNode(tmConf, privValidator, tmGenesisDoc, app, metricsProvider, tmLogger)
	if err != nil {
		return nil, err
	}

	transactor := execution.NewTransactor(kern.Blockchain, kern.Emitter, execution.NewAccounts(checker, keyClient, AccountsRingMutexCount),
		kern.Node.MempoolReactor().BroadcastTx, txCodec, kern.Logger)

	nameRegState := kern.State
	accountState := kern.State
	nodeView, err := tendermint.NewNodeView(kern.Node, txCodec, kern.RunID)
	if err != nil {
		return nil, err
	}
	kern.Service = rpc.NewService(accountState, nameRegState, kern.Blockchain, nodeView, kern.Logger)

	kern.Launchers = []process.Launcher{
		{
			Name:    "Profiling Server",
			Enabled: rpcConfig.Profiler.Enabled,
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
		},
		{
			Name:    "Database",
			Enabled: true,
			Launch: func() (process.Process, error) {
				// Just close database
				return process.ShutdownFunc(func(ctx context.Context) error {
					stateDB.Close()
					return nil
				}), nil
			},
		},
		{
			Name:    "Tendermint",
			Enabled: true,
			Launch: func() (process.Process, error) {
				err := kern.Node.Start()
				if err != nil {
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
		},
		{
			Name:    "RPC/info",
			Enabled: rpcConfig.Info.Enabled,
			Launch: func() (process.Process, error) {
				server, err := rpcinfo.StartServer(kern.Service, "/websocket", rpcConfig.Info.ListenAddress, kern.Logger)
				if err != nil {
					return nil, err
				}
				return server, nil
			},
		},
		{
			Name:    "RPC/metrics",
			Enabled: rpcConfig.Metrics.Enabled,
			Launch: func() (process.Process, error) {
				server, err := metrics.StartServer(kern.Service, rpcConfig.Metrics.MetricsPath,
					rpcConfig.Metrics.ListenAddress, rpcConfig.Metrics.BlockSampleSize, kern.Logger)
				if err != nil {
					return nil, err
				}
				return server, nil
			},
		},
		{
			Name:    "RPC/GRPC",
			Enabled: rpcConfig.GRPC.Enabled,
			Launch: func() (process.Process, error) {
				listen, err := net.Listen("tcp", rpcConfig.GRPC.ListenAddress)
				if err != nil {
					return nil, err
				}

				grpcServer := rpc.NewGRPCServer(kern.Logger)
				var ks *keys.KeyStore
				if keyStore != nil {
					ks = keyStore
				}

				if keyConfig.GRPCServiceEnabled {
					if keyStore == nil {
						ks = keys.NewKeyStore(keyConfig.KeysDirectory, keyConfig.AllowBadFilePermissions, kern.Logger)
					}
					keys.RegisterKeysServer(grpcServer, ks)
				}

				rpcquery.RegisterQueryServer(grpcServer, rpcquery.NewQueryServer(kern.State, nameRegState,
					kern.Blockchain, nodeView, kern.Logger))

				rpctransact.RegisterTransactServer(grpcServer, rpctransact.NewTransactServer(transactor, txCodec))

				rpcevents.RegisterExecutionEventsServer(grpcServer, rpcevents.NewExecutionEventsServer(kern.State,
					kern.Emitter, kern.Blockchain, kern.Logger))

				// Provides metadata about services registered
				//reflection.Register(grpcServer)

				go grpcServer.Serve(listen)

				return process.ShutdownFunc(func(ctx context.Context) error {
					grpcServer.Stop()
					// listener is closed for us
					return nil
				}), nil
			},
		},
	}

	return kern, nil
}

// Boot the kernel starting Tendermint and RPC layers
func (kern *Kernel) Boot() error {
	for _, launcher := range kern.Launchers {
		if launcher.Enabled {
			srvr, err := launcher.Launch()
			if err != nil {
				return fmt.Errorf("error launching %s server: %v", launcher.Name, err)
			}

			kern.processes[launcher.Name] = srvr
		}
	}
	go kern.supervise()
	return nil
}

func (kern *Kernel) Panic(err error) {
	fmt.Fprintf(os.Stderr, "%s: Kernel shutting down due to panic: %v", kern.nodeInfo, err)
	kern.Shutdown(context.Background())
	os.Exit(1)
}

// Wait for a graceful shutdown
func (kern *Kernel) WaitForShutdown() {
	// Supports multiple goroutines waiting for shutdown since channel is closed
	<-kern.shutdownNotify
}

// Supervise kernel once booted
func (kern *Kernel) supervise() {
	// perform disaster restarts of the kernel; rejoining the network as if we were a new node.
	shutdownCh := make(chan os.Signal, 1)
	reloadCh := make(chan os.Signal, 1)
	syncCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	signal.Notify(reloadCh, syscall.SIGHUP)
	signal.Notify(syncCh, syscall.SIGUSR1)
	for {
		select {
		case <-reloadCh:
			kern.Logger.Reload()
		case <-syncCh:
			kern.Logger.Sync()
		case sig := <-shutdownCh:
			kern.Logger.InfoMsg(fmt.Sprintf("Caught %v signal so shutting down", sig),
				"signal", sig.String())
			kern.Shutdown(context.Background())
			return
		}
	}
}

// Stop the kernel allowing for a graceful shutdown of components in order
func (kern *Kernel) Shutdown(ctx context.Context) (err error) {
	kern.shutdownOnce.Do(func() {
		logger := kern.Logger.WithScope("Shutdown")
		logger.InfoMsg("Attempting graceful shutdown...")
		logger.InfoMsg("Shutting down servers")
		ctx, cancel := context.WithTimeout(ctx, ServerShutdownTimeout)
		defer cancel()
		// Shutdown servers in reverse order to boot
		for i := len(kern.Launchers) - 1; i >= 0; i-- {
			name := kern.Launchers[i].Name
			srvr, ok := kern.processes[name]
			if ok {
				logger.InfoMsg("Shutting down server", "server_name", name)
				sErr := srvr.Shutdown(ctx)
				if sErr != nil {
					logger.InfoMsg("Failed to shutdown server",
						"server_name", name,
						structure.ErrorKey, sErr)
					if err == nil {
						err = sErr
					}
				}
			}
		}
		logger.InfoMsg("Shutdown complete")
		structure.Sync(kern.Logger.Info)
		structure.Sync(kern.Logger.Trace)
		// We don't want to wait for them, but yielding for a cooldown Let other goroutines flush
		// potentially interesting final output (e.g. log messages)
		time.Sleep(CooldownTime)
		close(kern.shutdownNotify)
	})
	return
}
