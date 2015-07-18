package erisdbss

import (
	"bytes"
	"encoding/json"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/gin-gonic/gin"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/binary"
	. "github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/common"
	"github.com/eris-ltd/eris-db/Godeps/_workspace/src/github.com/tendermint/tendermint/state"
	"github.com/eris-ltd/eris-db/server"
	"net/http"
	"os"
)

const TendermintConfigDefault = `# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

moniker = "__MONIKER__"
seeds = ""
fast_sync = false
db_backend = "leveldb"
log_level = "debug"
node_laddr = ""
`

// User data accepts a private validator and genesis json object.
// * PrivValidator is the private validator json data.
// * Genesis is the genesis json data.
// * MaxDuration is the maximum duration of the process (in seconds).
//   If this is 0, it will be set to REAPER_THRESHOLD
// TODO more stuff, like tendermint and server config files. Will probably
// wait with this until the eris/EPM integration.
type RequestData struct {
	PrivValidator *state.PrivValidator `json:"priv_validator"`
	Genesis       *state.GenesisDoc    `json:"genesis"`
	MaxDuration   uint                 `json:"max_duration"`
}

// The response is the port of the newly generated server. The assumption
// here is that the host name or ip is the same, and the default server
// settings apply.
// TODO return some "live" data after starting the node, so that
// the requester can validate that everything is fine. Maybe
// some data directly from the state manager. Genesis hash?
type ResponseData struct {
	Port string `json:"port"`
}

// Serves requests to fire up erisdb executables. POSTing to the server
// endpoint (/server by default) with RequestData in the body will create
// a fresh working directory with files based on that indata, fire up a
// new 'erisdb' executable and point it to that dir. The purpose is mostly
// to make testing easier, since setting up a node is as easy as making a
// http request.
// TODO link up with eris/EPM instead, to spawn new nodes in containers.
type ServerServer struct {
	running       bool
	serverManager *ServerManager
}

// Create a new ServerServer with the given base directory.
func NewServerServer(baseDir string) *ServerServer {
	os.RemoveAll(baseDir)
	EnsureDir(baseDir)
	return &ServerServer{serverManager: NewServerManager(100, baseDir)}
}

// Start the server.
func (this *ServerServer) Start(config *server.ServerConfig, router *gin.Engine) {
	router.POST("/server", this.handleFunc)
	this.running = true
}

// Is the server currently running.
func (this *ServerServer) Running() bool {
	return this.running
}

// Shut the server down. Will close all websocket sessions.
func (this *ServerServer) ShutDown() {
	this.running = false
	this.serverManager.killAll()
}

// Handle incoming requests.
func (this *ServerServer) handleFunc(c *gin.Context) {
	log.Debug("Incoming message")
	r := c.Request
	var buf bytes.Buffer
	n, errR := buf.ReadFrom(r.Body)
	if errR != nil || n == 0 {
		http.Error(c.Writer, "Bad request.", 400)
		return
	}
	bts := buf.Bytes()
	var errDC error
	reqData := &RequestData{}
	binary.ReadJSON(reqData, bts, &errDC)
	if errDC != nil {
		http.Error(c.Writer, "Failed to decode json.", 400)
		return
	}
	log.Debug("Starting to add.")
	resp, errA := this.serverManager.add(reqData)
	if errA != nil {
		http.Error(c.Writer, "Internal error: "+errA.Error(), 500)
		return
	}
	log.Debug("Work done.", "PORT", resp.Port)
	w := c.Writer
	enc := json.NewEncoder(w)
	enc.Encode(resp)
	w.WriteHeader(200)

}
