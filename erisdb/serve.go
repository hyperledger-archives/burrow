// The erisdb package contains tendermint-specific services that goes with the
// server.
package erisdb

import (
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/log15"
	. "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/common"
	cfg "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/config"
	tmcfg "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/config/tendermint"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/node"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/p2p"
	ep "github.com/eris-ltd/eris-db/erisdb/pipe"
	"github.com/eris-ltd/eris-db/server"
	"path"
)

const ERISDB_VERSION = "0.10.2"
const TENDERMINT_VERSION = "0.5.0"

var log = log15.New("module", "eris/erisdb_server")
var tmConfig cfg.Config

// This function returns a properly configured ErisDb server process with a running
// tendermint node attached to it. To start listening for incoming requests, call
// 'Start()' on the process. Make sure to register any start event listeners before
// that.
func ServeErisDB(workDir string) (*server.ServeProcess, error) {
	log.Info("ErisDB Serve initializing.")
	errEns := EnsureDir(workDir)

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
	tmConfig = tmcfg.GetConfig(workDir)
	tmConfig.Set("version", TENDERMINT_VERSION)
	cfg.ApplyConfig(tmConfig) // Notify modules of new config

	// Set the node up.
	nodeRd := make(chan struct{})
	nd := node.NewNode()
	// Load the supporting objects.
	pipe := ep.NewPipe(nd)
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

	stopChan := proc.StopEventChannel()
	go startNode(nd, nodeRd, stopChan)
	<-nodeRd
	return proc, nil
}

// Private. Create a new node
func startNode(nd *node.Node, ready chan struct{}, shutDown <-chan struct{}) {
	laddr := tmConfig.GetString("node_laddr")
	if laddr != "" {
		l := p2p.NewDefaultListener("tcp", laddr, false)
		nd.AddListener(l)
	}

	nd.Start()

	// If seedNode is provided by config, dial out.

	if len(tmConfig.GetString("seeds")) > 0 {
		nd.DialSeed()
	}

	ready <- struct{}{}
	// Block until everything is shut down.
	<-shutDown
	nd.Stop()
}
