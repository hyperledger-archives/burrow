package rpc_v0

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/eris-ltd/eris-db/blockchain"
	"github.com/eris-ltd/eris-db/core/pipes"
	core_types "github.com/eris-ltd/eris-db/core/types"
	definitions "github.com/eris-ltd/eris-db/definitions"
	event "github.com/eris-ltd/eris-db/event"
	rpc "github.com/eris-ltd/eris-db/rpc"
	server "github.com/eris-ltd/eris-db/server"
	"github.com/eris-ltd/eris-db/txs"
	"github.com/eris-ltd/eris-db/util"
)

// Provides a REST-like web-api. Implements server.Server
// TODO more routers. Also, start looking into how better status codes
// can be gotten.
type RestServer struct {
	codec         rpc.Codec
	pipe          definitions.Pipe
	eventSubs     *event.EventSubscriptions
	filterFactory *event.FilterFactory
	running       bool
}

// Create a new rest server.
func NewRestServer(codec rpc.Codec, pipe definitions.Pipe,
	eventSubs *event.EventSubscriptions) *RestServer {
	return &RestServer{
		codec:         codec,
		pipe:          pipe,
		eventSubs:     eventSubs,
		filterFactory: blockchain.NewBlockchainFilterFactory(),
	}
}

// Starting the server means registering all the handlers with the router.
func (restServer *RestServer) Start(config *server.ServerConfig, router *gin.Engine) {
	// Accounts
	router.GET("/accounts", parseSearchQuery, restServer.handleAccounts)
	router.GET("/accounts/:address", addressParam, restServer.handleAccount)
	router.GET("/accounts/:address/storage", addressParam, restServer.handleStorage)
	router.GET("/accounts/:address/storage/:key", addressParam, keyParam, restServer.handleStorageAt)
	// Blockchain
	router.GET("/blockchain", restServer.handleBlockchainInfo)
	router.GET("/blockchain/chain_id", restServer.handleChainId)
	router.GET("/blockchain/genesis_hash", restServer.handleGenesisHash)
	router.GET("/blockchain/latest_block_height", restServer.handleLatestBlockHeight)
	router.GET("/blockchain/latest_block", restServer.handleLatestBlock)
	router.GET("/blockchain/blocks", parseSearchQuery, restServer.handleBlocks)
	router.GET("/blockchain/block/:height", heightParam, restServer.handleBlock)
	// Consensus
	router.GET("/consensus", restServer.handleConsensusState)
	router.GET("/consensus/validators", restServer.handleValidatorList)
	// Events
	router.POST("/event_subs", restServer.handleEventSubscribe)
	router.GET("/event_subs/:id", restServer.handleEventPoll)
	router.DELETE("/event_subs/:id", restServer.handleEventUnsubscribe)
	// NameReg
	router.GET("/namereg", parseSearchQuery, restServer.handleNameRegEntries)
	router.GET("/namereg/:key", nameParam, restServer.handleNameRegEntry)
	// Network
	router.GET("/network", restServer.handleNetworkInfo)
	router.GET("/network/client_version", restServer.handleClientVersion)
	router.GET("/network/moniker", restServer.handleMoniker)
	router.GET("/network/listening", restServer.handleListening)
	router.GET("/network/listeners", restServer.handleListeners)
	router.GET("/network/peers", restServer.handlePeers)
	router.GET("/network/peers/:address", peerAddressParam, restServer.handlePeer)
	// Tx related (TODO get txs has still not been implemented)
	router.POST("/txpool", restServer.handleBroadcastTx)
	router.GET("/txpool", restServer.handleUnconfirmedTxs)
	// Code execution
	router.POST("/calls", restServer.handleCall)
	router.POST("/codecalls", restServer.handleCallCode)
	// Unsafe
	router.GET("/unsafe/pa_generator", restServer.handleGenPrivAcc)
	router.POST("/unsafe/txpool", parseTxModifier, restServer.handleTransact)
	router.POST("/unsafe/namereg/txpool", restServer.handleTransactNameReg)
	router.POST("/unsafe/tx_signer", restServer.handleSignTx)
	restServer.running = true
}

// Is the server currently running?
func (restServer *RestServer) Running() bool {
	return restServer.running
}

// Shut the server down. Does nothing.
func (restServer *RestServer) ShutDown() {
	restServer.running = false
}

// ********************************* Accounts *********************************

func (restServer *RestServer) handleGenPrivAcc(c *gin.Context) {
	addr := &AddressParam{}

	var acc interface{}
	var err error
	if addr.Address == nil || len(addr.Address) == 0 {
		acc, err = restServer.pipe.Accounts().GenPrivAccount()
	} else {
		acc, err = restServer.pipe.Accounts().GenPrivAccountFromKey(addr.Address)
	}
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(acc, c.Writer)
}

func (restServer *RestServer) handleAccounts(c *gin.Context) {
	var filters []*event.FilterData
	fs, exists := c.Get("filters")
	if exists {
		filters = fs.([]*event.FilterData)
	}
	accs, err := restServer.pipe.Accounts().Accounts(filters)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(accs, c.Writer)
}

func (restServer *RestServer) handleAccount(c *gin.Context) {
	addr := c.MustGet("addrBts").([]byte)
	acc, err := restServer.pipe.Accounts().Account(addr)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(acc, c.Writer)
}

func (restServer *RestServer) handleStorage(c *gin.Context) {
	addr := c.MustGet("addrBts").([]byte)
	s, err := restServer.pipe.Accounts().Storage(addr)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(s, c.Writer)
}

func (restServer *RestServer) handleStorageAt(c *gin.Context) {
	addr := c.MustGet("addrBts").([]byte)
	key := c.MustGet("keyBts").([]byte)
	sa, err := restServer.pipe.Accounts().StorageAt(addr, key)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(sa, c.Writer)
}

// ********************************* Blockchain *********************************

func (restServer *RestServer) handleBlockchainInfo(c *gin.Context) {
	bci := pipes.BlockchainInfo(restServer.pipe)
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(bci, c.Writer)
}

func (restServer *RestServer) handleGenesisHash(c *gin.Context) {
	gh := restServer.pipe.GenesisHash()
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(&core_types.GenesisHash{gh}, c.Writer)
}

func (restServer *RestServer) handleChainId(c *gin.Context) {
	cId := restServer.pipe.Blockchain().ChainId()
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(&core_types.ChainId{cId}, c.Writer)
}

func (restServer *RestServer) handleLatestBlockHeight(c *gin.Context) {
	lbh := restServer.pipe.Blockchain().Height()
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(&core_types.LatestBlockHeight{lbh}, c.Writer)
}

func (restServer *RestServer) handleLatestBlock(c *gin.Context) {
	latestHeight := restServer.pipe.Blockchain().Height()
	lb := restServer.pipe.Blockchain().Block(latestHeight)
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(lb, c.Writer)
}

func (restServer *RestServer) handleBlocks(c *gin.Context) {
	var filters []*event.FilterData
	fs, exists := c.Get("filters")
	if exists {
		filters = fs.([]*event.FilterData)
	}

	blocks, err := blockchain.FilterBlocks(restServer.pipe.Blockchain(),
		restServer.filterFactory, filters)

	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(blocks, c.Writer)
}

func (restServer *RestServer) handleBlock(c *gin.Context) {
	height := c.MustGet("height").(int)
	block := restServer.pipe.Blockchain().Block(height)
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(block, c.Writer)
}

// ********************************* Consensus *********************************
func (restServer *RestServer) handleConsensusState(c *gin.Context) {
	cs := restServer.pipe.GetConsensusEngine().ConsensusState()
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(cs, c.Writer)
}

func (restServer *RestServer) handleValidatorList(c *gin.Context) {
	vl := restServer.pipe.GetConsensusEngine().ListValidators()
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(vl, c.Writer)
}

// ********************************* Events *********************************

func (restServer *RestServer) handleEventSubscribe(c *gin.Context) {
	param := &EventIdParam{}
	errD := restServer.codec.Decode(param, c.Request.Body)
	if errD != nil {
		c.AbortWithError(500, errD)
	}
	subId, err := restServer.eventSubs.Add(param.EventId)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(&event.EventSub{subId}, c.Writer)
}

func (restServer *RestServer) handleEventPoll(c *gin.Context) {
	subId := c.MustGet("id").(string)
	data, err := restServer.eventSubs.Poll(subId)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(&event.PollResponse{data}, c.Writer)
}

func (restServer *RestServer) handleEventUnsubscribe(c *gin.Context) {
	subId := c.MustGet("id").(string)
	err := restServer.eventSubs.Remove(subId)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(&event.EventUnsub{true}, c.Writer)
}

// ********************************* NameReg *********************************

func (restServer *RestServer) handleNameRegEntries(c *gin.Context) {
	var filters []*event.FilterData
	fs, exists := c.Get("filters")
	if exists {
		filters = fs.([]*event.FilterData)
	}
	entries, err := restServer.pipe.NameReg().Entries(filters)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(entries, c.Writer)
}

func (restServer *RestServer) handleNameRegEntry(c *gin.Context) {
	name := c.MustGet("name").(string)
	entry, err := restServer.pipe.NameReg().Entry(name)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(entry, c.Writer)
}

// ********************************* Network *********************************

func (restServer *RestServer) handleNetworkInfo(c *gin.Context) {
	nInfo, err := pipes.NetInfo(restServer.pipe)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(nInfo, c.Writer)
}

func (restServer *RestServer) handleClientVersion(c *gin.Context) {
	version, err := pipes.ClientVersion(restServer.pipe)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(&core_types.ClientVersion{version}, c.Writer)
}

func (restServer *RestServer) handleMoniker(c *gin.Context) {
	moniker, err := pipes.Moniker(restServer.pipe)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(&core_types.Moniker{moniker}, c.Writer)
}

func (restServer *RestServer) handleListening(c *gin.Context) {
	listening, err := pipes.Listening(restServer.pipe)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(&core_types.Listening{listening}, c.Writer)
}

func (restServer *RestServer) handleListeners(c *gin.Context) {
	listeners, err := pipes.Listeners(restServer.pipe)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(&core_types.Listeners{listeners}, c.Writer)
}

func (restServer *RestServer) handlePeers(c *gin.Context) {
	peers, err := pipes.Peers(restServer.pipe)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(peers, c.Writer)
}

func (restServer *RestServer) handlePeer(c *gin.Context) {
	address := c.MustGet("address").(string)
	peer, err := pipes.Peer(restServer.pipe, address)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(peer, c.Writer)
}

// ********************************* Transactions *********************************

func (restServer *RestServer) handleBroadcastTx(c *gin.Context) {
	param := &txs.CallTx{}
	errD := restServer.codec.Decode(param, c.Request.Body)
	if errD != nil {
		c.AbortWithError(500, errD)
	}
	receipt, err := restServer.pipe.Transactor().BroadcastTx(param)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(receipt, c.Writer)
}

func (restServer *RestServer) handleUnconfirmedTxs(c *gin.Context) {
	trans, err := restServer.pipe.GetConsensusEngine().ListUnconfirmedTxs(-1)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(txs.UnconfirmedTxs{trans}, c.Writer)
}

func (restServer *RestServer) handleCall(c *gin.Context) {
	param := &CallParam{}
	errD := restServer.codec.Decode(param, c.Request.Body)
	if errD != nil {
		c.AbortWithError(500, errD)
	}
	call, err := restServer.pipe.Transactor().Call(param.From, param.Address, param.Data)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(call, c.Writer)
}

func (restServer *RestServer) handleCallCode(c *gin.Context) {
	param := &CallCodeParam{}
	errD := restServer.codec.Decode(param, c.Request.Body)
	if errD != nil {
		c.AbortWithError(500, errD)
	}
	call, err := restServer.pipe.Transactor().CallCode(param.From, param.Code, param.Data)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(call, c.Writer)
}

func (restServer *RestServer) handleTransact(c *gin.Context) {

	_, hold := c.Get("hold")

	param := &TransactParam{}
	errD := restServer.codec.Decode(param, c.Request.Body)
	if errD != nil {
		c.AbortWithError(500, errD)
	}
	if hold {
		res, err := restServer.pipe.Transactor().TransactAndHold(param.PrivKey, param.Address, param.Data, param.GasLimit, param.Fee)
		if err != nil {
			c.AbortWithError(500, err)
		}
		c.Writer.WriteHeader(200)
		restServer.codec.Encode(res, c.Writer)
	} else {
		receipt, err := restServer.pipe.Transactor().Transact(param.PrivKey, param.Address, param.Data, param.GasLimit, param.Fee)
		if err != nil {
			c.AbortWithError(500, err)
		}
		c.Writer.WriteHeader(200)
		restServer.codec.Encode(receipt, c.Writer)
	}
}

func (restServer *RestServer) handleTransactNameReg(c *gin.Context) {
	param := &TransactNameRegParam{}
	errD := restServer.codec.Decode(param, c.Request.Body)
	if errD != nil {
		c.AbortWithError(500, errD)
	}
	receipt, err := restServer.pipe.Transactor().TransactNameReg(param.PrivKey, param.Name, param.Data, param.Amount, param.Fee)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(receipt, c.Writer)
}

func (restServer *RestServer) handleSignTx(c *gin.Context) {
	param := &SignTxParam{}
	errD := restServer.codec.Decode(param, c.Request.Body)
	if errD != nil {
		c.AbortWithError(500, errD)
	}
	tx, err := restServer.pipe.Transactor().SignTx(param.Tx, param.PrivAccounts)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	restServer.codec.Encode(tx, c.Writer)
}

// ********************************* Middleware *********************************

func addressParam(c *gin.Context) {
	addr := c.Param("address")
	if !util.IsAddress(addr) {
		c.AbortWithError(400, fmt.Errorf("Malformed address param: "+addr))
	}
	bts, _ := hex.DecodeString(addr)
	c.Set("addrBts", bts)
	c.Next()
}

func nameParam(c *gin.Context) {
	name := c.Param("key")
	c.Set("name", name)
	c.Next()
}

func keyParam(c *gin.Context) {
	key := c.Param("key")
	bts, err := hex.DecodeString(key)
	if err != nil {
		c.AbortWithError(400, err)
	}
	c.Set("keyBts", bts)
	c.Next()
}

func heightParam(c *gin.Context) {
	h, err := strconv.Atoi(c.Param("height"))
	if err != nil {
		c.AbortWithError(400, err)
	}
	if h < 0 {
		c.AbortWithError(400, fmt.Errorf("Negative number used as height."))
	}
	c.Set("height", h)
	c.Next()
}

func subIdParam(c *gin.Context) {
	subId := c.Param("id")
	if len(subId) != 64 || !util.IsHex(subId) {
		c.AbortWithError(400, fmt.Errorf("Malformed event id"))
	}
	c.Set("id", subId)
	c.Next()
}

// TODO
func peerAddressParam(c *gin.Context) {
	subId := c.Param("address")
	c.Set("address", subId)
	c.Next()
}

func parseTxModifier(c *gin.Context) {
	hold := c.Query("hold")
	if hold == "true" {
		c.Set("hold", true)
	} else if hold != "" {
		if hold != "false" {
			c.Writer.WriteHeader(400)
			c.Writer.Write([]byte("tx hold must be either 'true' or 'false', found: " + hold))
			c.Abort()
		}
	}
}

func parseSearchQuery(c *gin.Context) {
	q := c.Query("q")
	if q != "" {
		data, err := _parseSearchQuery(q)
		if err != nil {
			c.Writer.WriteHeader(400)
			c.Writer.Write([]byte(err.Error()))
			c.Abort()
			// c.AbortWithError(400, err)
			return
		}
		c.Set("filters", data)
	}
}

func _parseSearchQuery(queryString string) ([]*event.FilterData, error) {
	if len(queryString) == 0 {
		return nil, nil
	}
	filters := strings.Split(queryString, " ")
	fdArr := []*event.FilterData{}
	for _, f := range filters {
		kv := strings.Split(f, ":")
		if len(kv) != 2 {
			return nil, fmt.Errorf("Malformed query. Missing ':' separator: " + f)
		}
		if kv[0] == "" {
			return nil, fmt.Errorf("Malformed query. Field name missing: " + f)
		}

		fd, fd2, errTfd := toFilterData(kv[0], kv[1])
		if errTfd != nil {
			return nil, errTfd
		}
		fdArr = append(fdArr, fd)
		if fd2 != nil {
			fdArr = append(fdArr, fd2)
		}
	}
	return fdArr, nil
}

// Parse the query statement and create . Two filter data in case of a range param.
func toFilterData(field, stmt string) (*event.FilterData, *event.FilterData, error) {
	// In case statement is empty
	if stmt == "" {
		return &event.FilterData{field, "==", ""}, nil, nil
	}
	// Simple routine based on string splitting. TODO add quoted range query.
	if stmt[0] == '>' || stmt[0] == '<' || stmt[0] == '=' || stmt[0] == '!' {
		// restServer means a normal operator. If one character then stop, otherwise
		// peek at next and check if it's a "=".

		if len(stmt) == 1 {
			return &event.FilterData{field, stmt[0:1], ""}, nil, nil
		} else if stmt[1] == '=' {
			return &event.FilterData{field, stmt[:2], stmt[2:]}, nil, nil
		} else {
			return &event.FilterData{field, stmt[0:1], stmt[1:]}, nil, nil
		}
	} else {
		// Either we have a range query here or a malformed query.
		rng := strings.Split(stmt, "..")
		// restServer is for when there is no op, but the value is not empty.
		if len(rng) == 1 {
			return &event.FilterData{field, "==", stmt}, nil, nil
		}
		// The rest.
		if len(rng) != 2 || rng[0] == "" || rng[1] == "" {
			return nil, nil, fmt.Errorf("Malformed query statement: " + stmt)
		}
		var min string
		var max string
		if rng[0] == "*" {
			min = "min"
		} else {
			min = rng[0]
		}
		if rng[1] == "*" {
			max = "max"
		} else {
			max = rng[1]
		}
		return &event.FilterData{field, ">=", min}, &event.FilterData{field, "<=", max}, nil
	}
	return nil, nil, nil
}
