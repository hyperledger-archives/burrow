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
// To start listening for incoming HTTP requests on the Rest server, call 'Start()' on the process.
// Make sure to register any start event listeners first
func ServeErisDB(workDir string, inProc bool) (*server.ServeProcess, error) {
	log.Info("ErisDB Serve initializing.")
	errEns := EnsureDir(workDir, 0777)

	if errEns != nil {
		return nil, errEns
	}

	// there are two types of config we need to load,
	// one for the erisdb server and one for tendermint.
	// even if consensus isn't in process, the tendermint libs (eg. db)
	// expect tendermint/go-config to be setup.
	// Regardless, both configs are expected in the same file (root/config.toml)
	// Some of this stuff is implicit and maybe a little confusing,
	// but cfg mgmt across projects probably often is!

	// Get an erisdb configuration
	var sConf *server.ServerConfig
	sConfPath := path.Join(workDir, "config.toml")
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
	tmConfig = tmcfg.GetConfig(workDir)
	// tmConfig.Set("tm.version", TENDERMINT_VERSION) // ?
	cfg.ApplyConfig(tmConfig) // Notify in proc tendermint of new config

	// Load the application state
	// The app state used to be managed by tendermint node,
	// but is now managed by ErisDB.
	// The tendermint core only stores the blockchain (history of txs)
	stateDB := dbm.GetDB("app_state", config.GetString("erisdb.db_backend"), config.GetString("erisdb.db_dir"))
	state := sm.LoadState(stateDB)
	var genDoc *stypes.GenesisDoc
	if state == nil {
		genDoc, state = sm.MakeGenesisStateFromFile(stateDB, config.GetString("genesis_file"))
		state.Save()
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
	// for now it's same for both of them
	config.Set("tm.chain_id", state.ChainID)
	config.Set("erisdb.chain_id", state.ChainID)

	// *****************************
	// erisdb-tmsp app

	// start the event switch for state related events
	// (transactions to/from acconts, etc)
	evsw := events.NewEventSwitch()
	evsw.Start()

	// create the app
	app := edbapp.NewErisDBApp(state, evsw)

	// so we know where to find the consensus host (for eg. blockchain/consensus rpcs)
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

	// *****************************
	// Boot the erisdb restful API servers

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

	return proc, nil
}

// start an inproc tendermint node
func startTMNode(app *edbapp.ErisDBApp) {
	// get the genesis
	genDocFile := config.GetString("tm.genesis_file")
	jsonBlob, err := ioutil.ReadFile(genDocFile)
	if err != nil {
		Exit(Fmt("Couldn't read GenesisDoc file: %v", err))
	}
	genDoc := types.GenesisDocFromJSON(jsonBlob)
	if genDoc.ChainID == "" {
		PanicSanity(Fmt("Genesis doc %v must include non-empty chain_id", genDocFile))
	}
	config.Set("tm.chain_id", genDoc.ChainID)
	config.Set("tm.genesis_doc", genDoc)

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
