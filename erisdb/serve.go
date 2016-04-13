// The erisdb package contains tendermint-specific services that goes with the
// server.
package erisdb

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"sync"

	. "github.com/tendermint/go-common"
	cfg "github.com/tendermint/go-config"
	dbm "github.com/tendermint/go-db"
	"github.com/tendermint/go-events"
	"github.com/tendermint/go-p2p"
	"github.com/tendermint/go-wire"
	"github.com/tendermint/log15"

	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/types"
	tmspcli "github.com/tendermint/tmsp/client"
	tmsp "github.com/tendermint/tmsp/server"

	tmcfg "github.com/eris-ltd/eris-db/config/tendermint"
	ep "github.com/eris-ltd/eris-db/erisdb/pipe"
	"github.com/eris-ltd/eris-db/server"
	sm "github.com/eris-ltd/eris-db/state"
	stypes "github.com/eris-ltd/eris-db/state/types"
	edbapp "github.com/eris-ltd/eris-db/tmsp"
)

const ERISDB_VERSION = "0.11.5"
const TENDERMINT_VERSION = "0.5.0"

var log = log15.New("module", "eris/erisdb_server")
var tmConfig cfg.Config

// This function returns a properly configured ErisDb server process,
// with a tmsp listener for talking to tendermint core.
// To start listening for incoming requests, call 'Start()' on the process.
// Make sure to register any start event listeners first
func ServeErisDB(workDir string, inProc bool) (*server.ServeProcess, error) {
	log.Info("ErisDB Serve initializing.")
	errEns := EnsureDir(workDir, 0777)

	if errEns != nil {
		return nil, errEns
	}

	var sConf *server.ServerConfig

	sConfPath := path.Join(workDir, "server_conf.toml")
	if !FileExists(sConfPath) {
		log.Info("No server configuration, using default.")
		log.Info("Writing to: " + sConfPath)
		sConf = server.DefaultServerConfig()
		errW := server.WriteServerConfig(sConfPath, sConf)
		if errW != nil {
			panic(errW)
		}
	} else {
		var errRSC error
		sConf, errRSC = server.ReadServerConfig(sConfPath)
		if errRSC != nil {
			log.Error("Server config file error.", "error", errRSC.Error())
		}
	}

	// Get tendermint configuration
	// TODO replace
	tmConfig = tmcfg.GetConfig(workDir)
	tmConfig.Set("version", TENDERMINT_VERSION)
	cfg.ApplyConfig(tmConfig) // Notify modules of new config

	// Load the application state
	// The app state used to be managed by tendermint node,
	// but is now managed by ErisDB.
	// The tendermint core only stores the blockchain (history of txs)
	stateDB := dbm.GetDB("app_state")
	state := sm.LoadState(stateDB)
	var genDoc *stypes.GenesisDoc
	if state == nil {
		genDoc, state = sm.MakeGenesisStateFromFile(stateDB, config.GetString("erisdb_genesis_file"))
		state.Save()
		// write the gendoc to db
		buf, n, err := new(bytes.Buffer), new(int), new(error)
		wire.WriteJSON(genDoc, buf, n, err)
		stateDB.Set(stypes.GenDocKey, buf.Bytes())
		if *err != nil {
			Exit(Fmt("Unable to write gendoc to db: %v", err))
		}
	} else {
		genDocBytes := stateDB.Get(stypes.GenDocKey)
		err := new(error)
		wire.ReadJSONPtr(&genDoc, genDocBytes, err)
		if *err != nil {
			Exit(Fmt("Unable to read gendoc from db: %v", err))
		}
	}
	// add the chainid to the global config
	config.Set("chain_id", state.ChainID)

	evsw := events.NewEventSwitch()
	evsw.Start()

	app := edbapp.NewErisDBApp(state, evsw)
	app.SetHostAddress(sConf.Consensus.TendermintHost)

	if inProc {
		fmt.Println("Starting tm node in proc")
		startTMNode(app)
	} else {
		fmt.Println("Starting tmsp listener")
		// Start the tmsp listener for state update commands
		go func() {
			// TODO config
			_, err := tmsp.NewServer(sConf.Consensus.TMSPListener, app)
			if err != nil {
				// TODO: play nice
				Exit(err.Error())
			}
		}()
	}

	// Load supporting objects.
	pipe := ep.NewPipe(app, evsw)
	codec := &TCodec{}
	evtSubs := NewEventSubscriptions(pipe.Events())
	// The services.
	tmwss := NewErisDbWsService(codec, pipe)
	tmjs := NewErisDbJsonService(codec, pipe, evtSubs)
	// The servers.
	jsonServer := NewJsonRpcServer(tmjs)
	restServer := NewRestServer(codec, pipe, evtSubs)
	wsServer := server.NewWebSocketServer(sConf.WebSocket.MaxWebSocketSessions, tmwss)
	// Create a server process.
	proc := server.NewServeProcess(sConf, jsonServer, restServer, wsServer)

	//stopChan := proc.StopEventChannel()
	//go startNode(nd, nodeRd, stopChan)
	//<-nodeRd
	return proc, nil
}

func startTMNode(app *edbapp.ErisDBApp) {
	// get the genesis
	genDocFile := config.GetString("tendermint_genesis_file")
	jsonBlob, err := ioutil.ReadFile(genDocFile)
	if err != nil {
		Exit(Fmt("Couldn't read GenesisDoc file: %v", err))
	}
	genDoc := types.GenesisDocFromJSON(jsonBlob)
	if genDoc.ChainID == "" {
		PanicSanity(Fmt("Genesis doc %v must include non-empty chain_id", genDocFile))
	}
	config.Set("chain_id", genDoc.ChainID)
	config.Set("genesis_doc", genDoc)

	// Get PrivValidator
	privValidatorFile := config.GetString("priv_validator_file")
	privValidator := types.LoadOrGenPrivValidator(privValidatorFile)
	nd := node.NewNode(privValidator, func(addr string, hash []byte) proxy.AppConn {
		// TODO: Check the hash
		return tmspcli.NewLocalClient(new(sync.Mutex), app)
	})

	l := p2p.NewDefaultListener("tcp", config.GetString("node_laddr"), config.GetBool("skip_upnp"))
	nd.AddListener(l)
	if err := nd.Start(); err != nil {
		Exit(Fmt("Failed to start node: %v", err))
	}

	log.Notice("Started node", "nodeInfo", nd.NodeInfo())

	// If seedNode is provided by config, dial out.
	if config.GetString("seeds") != "" {
		seeds := strings.Split(config.GetString("seeds"), ",")
		nd.DialSeeds(seeds)
	}

	// Run the RPC server.
	if config.GetString("rpc_laddr") != "" {
		_, err := nd.StartRPC()
		if err != nil {
			PanicCrisis(err)
		}
	}
}

// Private. Create a new node.
func startNode(nd *node.Node, ready chan struct{}, shutDown <-chan struct{}) {
	laddr := tmConfig.GetString("node_laddr")
	if laddr != "" {
		l := p2p.NewDefaultListener("tcp", laddr, tmConfig.GetBool("skip_upnp"))
		nd.AddListener(l)
	}

	nd.Start()

	/*
			// If seedNode is provided by config, dial out.
			// should be handled by core

		if len(tmConfig.GetString("seeds")) > 0 {
				nd.DialSeed()
			}*/

	if len(tmConfig.GetString("rpc_laddr")) > 0 {
		nd.StartRPC()
	}
	ready <- struct{}{}
	// Block until everything is shut down.
	<-shutDown
	nd.Stop()
}
