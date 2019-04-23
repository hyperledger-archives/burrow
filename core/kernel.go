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
	"bytes"
	"context"
	"fmt"
	"net"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/hyperledger/burrow/bcm"
	"github.com/hyperledger/burrow/consensus/tendermint"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution"
	"github.com/hyperledger/burrow/execution/state"
	"github.com/hyperledger/burrow/genesis"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/logging/structure"
	"github.com/hyperledger/burrow/process"
	"github.com/hyperledger/burrow/rpc"
	"github.com/hyperledger/burrow/txs"
	"github.com/streadway/simpleuuid"
	"github.com/tendermint/tendermint/blockchain"
	dbm "github.com/tendermint/tendermint/libs/db"
	tmTypes "github.com/tendermint/tendermint/types"
)

const (
	CooldownTime           = 1000 * time.Millisecond
	ServerShutdownTimeout  = 5000 * time.Millisecond
	LoggingCallerDepth     = 5
	AccountsRingMutexCount = 100
	BurrowDBName           = "burrow_state"
)

// Kernel is the root structure of Burrow
type Kernel struct {
	// Expose these public-facing interfaces to allow programmatic extension of the Kernel by other projects
	Emitter        *event.Emitter
	Service        *rpc.Service
	Launchers      []process.Launcher
	State          *state.State
	Blockchain     *bcm.Blockchain
	Node           *tendermint.Node
	Transactor     *execution.Transactor
	RunID          simpleuuid.UUID // Time-based UUID randomly generated each time Burrow is started
	Logger         *logging.Logger
	database       dbm.DB
	txCodec        txs.Codec
	exeOptions     []execution.ExecutionOption
	checker        execution.BatchExecutor
	committer      execution.BatchCommitter
	keyClient      keys.KeyClient
	keyStore       *keys.KeyStore
	info           string
	processes      map[string]process.Process
	listeners      map[string]net.Listener
	timeoutFactor  float64
	shutdownNotify chan struct{}
	shutdownOnce   sync.Once
}

// NewKernel initializes an empty kernel
func NewKernel(dbDir string) (*Kernel, error) {
	if dbDir == "" {
		return nil, fmt.Errorf("Burrow requires a database directory")
	}
	runID, err := simpleuuid.NewTime(time.Now()) // Create a random ID based on start time
	return &Kernel{
		Logger:         logging.NewNoopLogger(),
		RunID:          runID,
		Emitter:        event.NewEmitter(),
		processes:      make(map[string]process.Process),
		listeners:      make(map[string]net.Listener),
		shutdownNotify: make(chan struct{}),
		txCodec:        txs.NewAminoCodec(),
		database:       dbm.NewDB(BurrowDBName, dbm.GoLevelDBBackend, dbDir),
	}, err
}

// SetLogger initializes the kernel with the provided logger
func (kern *Kernel) SetLogger(logger *logging.Logger) {
	logger = logger.WithScope("NewKernel()").With(structure.TimeKey,
		log.DefaultTimestampUTC, structure.RunId, kern.RunID.String())
	heightValuer := log.Valuer(func() interface{} { return kern.Blockchain.LastBlockHeight() })
	kern.Logger = logger.WithInfo(structure.CallerKey, log.Caller(LoggingCallerDepth)).With("height", heightValuer)
	kern.Emitter.SetLogger(logger)
}

// LoadState starts from scratch or previous chain
func (kern *Kernel) LoadState(genesisDoc *genesis.GenesisDoc) (err error) {
	var existing bool
	existing, kern.Blockchain, err = bcm.LoadOrNewBlockchain(kern.database, genesisDoc, kern.Logger)
	if err != nil {
		return fmt.Errorf("error creating or loading blockchain state: %v", err)
	}

	if existing {
		kern.Logger.InfoMsg("Loading application state", "height", kern.Blockchain.LastBlockHeight())
		kern.State, err = state.LoadState(kern.database, execution.VersionAtHeight(kern.Blockchain.LastBlockHeight()))
		if err != nil {
			return fmt.Errorf("could not load persisted execution state at hash 0x%X: %v",
				kern.Blockchain.AppHashAfterLastBlock(), err)
		}

		if !bytes.Equal(kern.State.Hash(), kern.Blockchain.AppHashAfterLastBlock()) {
			return fmt.Errorf("state and blockchain disagree on app hash at height %d: "+
				"state gives %X, blockchain gives %X", kern.Blockchain.LastBlockHeight(),
				kern.State.Hash(), kern.Blockchain.AppHashAfterLastBlock())
		}

	} else {
		kern.Logger.InfoMsg("Creating new application state from genesis")
		kern.State, err = state.MakeGenesisState(kern.database, genesisDoc)
		if err != nil {
			return fmt.Errorf("could not build genesis state: %v", err)
		}

		if err = kern.State.InitialCommit(); err != nil {
			return err
		}
	}

	kern.Logger.InfoMsg("State loading successful")

	params := execution.ParamsFromGenesis(genesisDoc)
	kern.checker = execution.NewBatchChecker(kern.State, params, kern.Blockchain, kern.Logger)
	kern.committer = execution.NewBatchCommitter(kern.State, params, kern.Blockchain, kern.Emitter, kern.Logger, kern.exeOptions...)
	return nil
}

// LoadDump restores chain state from the given dump file
func (kern *Kernel) LoadDump(genesisDoc *genesis.GenesisDoc, restoreFile string, silent bool) (err error) {
	var exists bool
	if exists, kern.Blockchain, err = bcm.LoadOrNewBlockchain(kern.database, genesisDoc, kern.Logger); err != nil {
		return fmt.Errorf("error creating or loading blockchain state: %v", err)
	}

	if exists {
		if silent {
			kern.Logger.InfoMsg("State already exists, skipping...")
			return nil
		}
		return fmt.Errorf("existing state found, please remove before restoring")
	}

	kern.Blockchain.SetBlockStore(bcm.NewBlockStore(blockchain.NewBlockStore(kern.database)))

	if kern.State, err = state.MakeGenesisState(kern.database, genesisDoc); err != nil {
		return fmt.Errorf("could not build genesis state: %v", err)
	}

	if len(genesisDoc.AppHash) == 0 {
		return fmt.Errorf("AppHash is required when restoring chain")
	}

	reader, err := state.NewFileDumpReader(restoreFile)
	if err != nil {
		return err
	}

	if err = kern.State.LoadDump(reader); err != nil {
		return err
	}

	if err = kern.State.InitialCommit(); err != nil {
		return err
	}

	if !bytes.Equal(kern.State.Hash(), kern.Blockchain.GenesisDoc().AppHash) {
		return fmt.Errorf("Restore produced a different apphash expect 0x%x got 0x%x",
			kern.Blockchain.GenesisDoc().AppHash, kern.State.Hash())
	}
	err = kern.Blockchain.CommitWithAppHash(kern.State.Hash())
	if err != nil {
		return fmt.Errorf("Unable to commit %v", err)
	}

	kern.Logger.InfoMsg("State restore successful -> height 0",
		"state_hash", kern.State.Hash())
	return nil
}

// GetNodeView builds and returns a wrapper of our tendermint node
func (kern *Kernel) GetNodeView() (*tendermint.NodeView, error) {
	if kern.Node == nil {
		return nil, nil
	}
	return tendermint.NewNodeView(kern.Node, kern.txCodec, kern.RunID)
}

// AddExecutionOptions extends our execution options
func (kern *Kernel) AddExecutionOptions(opts ...execution.ExecutionOption) {
	kern.exeOptions = append(kern.exeOptions, opts...)
}

// AddProcesses extends the services that we launch at boot
func (kern *Kernel) AddProcesses(pl ...process.Launcher) {
	kern.Launchers = append(kern.Launchers, pl...)
}

// SetKeyClient explicitly sets the key client
func (kern *Kernel) SetKeyClient(client keys.KeyClient) {
	kern.keyClient = client
}

// SetKeyStore explicitly sets the key store
func (kern *Kernel) SetKeyStore(store *keys.KeyStore) {
	kern.keyStore = store
}

// Generates an in-memory Tendermint PrivValidator (suitable for passing to LoadTendermintFromConfig)
func (kern *Kernel) PrivValidator(validator crypto.Address) (tmTypes.PrivValidator, error) {
	val, err := keys.AddressableSigner(kern.keyClient, validator)
	if err != nil {
		return nil, fmt.Errorf("could not get validator addressable from keys client: %v", err)
	}
	signer, err := keys.AddressableSigner(kern.keyClient, val.GetAddress())
	if err != nil {
		return nil, err
	}
	return tendermint.NewPrivValidatorMemory(val, signer), nil
}

// Boot the kernel starting Tendermint and RPC layers
func (kern *Kernel) Boot() (err error) {
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
	fmt.Fprintf(os.Stderr, "%v: shutting down due to panic: %v", kern, err)
	kern.ShutdownAndExit()
}

// Wait for a graceful shutdown
func (kern *Kernel) WaitForShutdown() {
	// Supports multiple goroutines waiting for shutdown since channel is closed
	<-kern.shutdownNotify
}

func (kern *Kernel) registerListener(name string, listener net.Listener) error {
	_, ok := kern.listeners[name]
	if ok {
		return fmt.Errorf("registerListener(): listener '%s' already registered", name)
	}
	kern.listeners[name] = listener
	return nil
}

func (kern *Kernel) GRPCListenAddress() net.Addr {
	l, ok := kern.listeners[GRPCProcessName]
	if !ok {
		return nil
	}
	return l.Addr()
}

func (kern *Kernel) InfoListenAddress() net.Addr {
	l, ok := kern.listeners[InfoProcessName]
	if !ok {
		return nil
	}
	return l.Addr()
}

func (kern *Kernel) MetricsListenAddress() net.Addr {
	l, ok := kern.listeners[MetricsProcessName]
	if !ok {
		return nil
	}
	return l.Addr()
}

func (kern *Kernel) String() string {
	return fmt.Sprintf("Kernel[%s]", kern.info)
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
			err := kern.Logger.Reload()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v: could not reload logger: %v", kern, err)
			}
		case <-syncCh:
			err := kern.Logger.Sync()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v: could not sync logger: %v", kern, err)
			}
		case sig := <-shutdownCh:
			kern.Logger.InfoMsg(fmt.Sprintf("Caught %v signal so shutting down", sig),
				"signal", sig.String())
			kern.ShutdownAndExit()
			return
		}
	}
}

func (kern *Kernel) ShutdownAndExit() {
	ctx, cancel := context.WithTimeout(context.Background(), ServerShutdownTimeout)
	defer cancel()
	err := kern.Shutdown(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: error shutting down: %v", kern, err)
		os.Exit(1)
	}
	os.Exit(0)
}

// Shutdown stops the kernel allowing for a graceful shutdown of components in order
func (kern *Kernel) Shutdown(ctx context.Context) (err error) {
	kern.shutdownOnce.Do(func() {
		logger := kern.Logger.WithScope("Shutdown")
		logger.InfoMsg("Attempting graceful shutdown...")
		logger.InfoMsg("Shutting down servers")
		// Shutdown servers in reverse order to boot
		for i := len(kern.Launchers) - 1; i >= 0; i-- {
			name := kern.Launchers[i].Name
			proc, ok := kern.processes[name]
			if ok {
				logger.InfoMsg("Shutting down server", "server_name", name)
				sErr := proc.Shutdown(ctx)
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
		// Best effort
		structure.Sync(kern.Logger.Info)
		structure.Sync(kern.Logger.Trace)
		// We don't want to wait for them, but yielding for a cooldown Let other goroutines flush
		// potentially interesting final output (e.g. log messages)
		time.Sleep(CooldownTime)
		close(kern.shutdownNotify)
	})
	return
}
