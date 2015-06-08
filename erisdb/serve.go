package erisdb

import (
	"fmt"
	"github.com/tendermint/log15"
	"github.com/eris-ltd/erisdb/server"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	ep "github.com/eris-ltd/erisdb/erisdb/pipe"
	. "github.com/tendermint/tendermint/common"
	cfg "github.com/tendermint/tendermint/config"
	"path"
	tmcfg "github.com/tendermint/tendermint/config/tendermint"
)

var log = log15.New("module", "eris/erisdb_server")
var tmConfig cfg.Config

func init() {
	cfg.OnConfig(func(newConfig cfg.Config) {
			fmt.Println("NEWCONFIG")
		tmConfig = newConfig
	})
}

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
		sConf = server.DefaultServerConfig()
		server.WriteServerConfig(sConfPath, sConf)
	} else {
		var errRSC error
		sConf, errRSC = server.ReadServerConfig(sConfPath)
		if errRSC != nil {
			log.Error("Server config file error.", "error", errRSC.Error())
		}
	}
	
	// Get tendermint configuration
	tmConfig = tmcfg.GetConfig(workDir)
	cfg.ApplyConfig(tmConfig) // Notify modules of new config
	
	// Set the node up.
	nodeRd := make(chan struct{})
	nd := node.NewNode()
	// Load the supporting objects.
	pipe := ep.NewPipe(nd)
	codec := &TCodec{}
	evtSubs := NewEventSubscriptions(pipe)
	// The services.
	tmwss := NewErisDbWsService(codec, pipe)
	tmjs := NewErisDbJsonService(codec, pipe, evtSubs)
	// The servers.
	jsonServer := NewJsonRpcServer(tmjs)
	restServer := NewRestServer(codec, pipe, evtSubs)
	wsServer := server.NewWebSocketServer(sConf.MaxWebSocketSessions, tmwss)
	// Create a server process.
	proc := server.NewServeProcess(sConf, jsonServer, restServer, wsServer)
	
	stopChan := proc.StopEventChannel()
	go startNode(nd, nodeRd, stopChan)
	<- nodeRd
	return proc, nil
}


// Private. Create a new node
func startNode(nd *node.Node, ready chan struct{}, shutDown <- chan struct{}) {
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
	<- shutDown
	nd.Stop()
}