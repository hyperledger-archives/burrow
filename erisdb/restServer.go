package erisdb

import (
	"encoding/hex"
	"fmt"
	ep "github.com/eris-ltd/erisdb/erisdb/pipe"
	rpc "github.com/eris-ltd/erisdb/rpc"
	"github.com/eris-ltd/erisdb/server"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

// Provides a REST-like web-api. Implements server.Server
// TODO more routers. Also, start looking into how better status codes
// can be gotten.
type RestServer struct {
	codec     rpc.Codec
	pipe      ep.Pipe
	eventSubs *EventSubscriptions
	running   bool
}

// Create a new rest server.
func NewRestServer(codec rpc.Codec, pipe ep.Pipe, eventSubs *EventSubscriptions) *RestServer {
	return &RestServer{codec: codec, pipe: pipe, eventSubs: eventSubs}
}

// Starting the server means registering all the handlers with the router.
func (this *RestServer) Start(config *server.ServerConfig, router *gin.Engine) {
	// Accounts
	router.GET("/accounts", this.handleAccounts)
	router.GET("/accounts/:address", addressParam, this.handleAccount)
	router.GET("/accounts/:address/storage", addressParam, this.handleStorage)
	router.GET("/accounts/:address/storage/:key", addressParam, keyParam, this.handleStorageAt)
	// Blockchain
	router.GET("/blockchain", this.handleBlockchainInfo)
	router.GET("/blockchain/chain_id", this.handleChainId)
	router.GET("/blockchain/genesis_hash", this.handleGenesisHash)
	router.GET("/blockchain/latest_block_height", this.handleLatestBlockHeight)
	router.GET("/blockchain/latest_block", this.handleLatestBlock)
	router.GET("/blockchain/blocks", this.handleBlocks)
	router.GET("/blockchain/block/:height", heightParam, this.handleBlock)
	// Consensus
	router.GET("/consensus/state", this.handleConsensusState)
	router.GET("/consensus/validators", this.handleValidatorList)
	// Events
	router.POST("/event_subs", this.handleEventSubscribe)
	router.GET("/event_subs/:id", this.handleEventPoll)
	router.DELETE("/event_subs/:id", this.handleEventUnsubscribe)
	// Network
	router.GET("/network", this.handleNetworkInfo)
	router.GET("/network/moniker", this.handleMoniker)
	router.GET("/network/listening", this.handleListening)
	router.GET("/network/listeners", this.handleListeners)
	router.GET("/network/peers", this.handlePeers)
	router.GET("/network/peers/:address", peerAddressParam, this.handlePeer)
	// Tx related (TODO get txs has still not been implemented)
	router.POST("/txpool", this.handleBroadcastTx)
	router.GET("/txpool", this.handleUnconfirmedTxs)
	// Code execution
	router.POST("/calls/:address", this.handleCall)
	router.POST("/calls/code", this.handleCallCode)
	// Unsafe
	router.GET("/unsafe/pa_generator", this.handleGenPrivAcc)
	router.POST("/unsafe/txpool", this.handleTransact)
	router.POST("/unsafe/tx_signer", this.handleSignTx)
	this.running = true
}

// Is the server currently running?
func (this *RestServer) Running() bool {
	return this.running
}

// Shut the server down. Does nothing.
func (this *RestServer) ShutDown() {
	this.running = false
}

// ********************************* Accounts *********************************

func (this *RestServer) handleGenPrivAcc(c *gin.Context) {
	addr := &AddressParam{}

	var acc interface{}
	var err error
	if addr.Address == nil || len(addr.Address) == 0 {
		acc, err = this.pipe.Accounts().GenPrivAccount()
	} else {
		acc, err = this.pipe.Accounts().GenPrivAccountFromKey(addr.Address)
	}
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(acc, c.Writer)
}

func (this *RestServer) handleAccounts(c *gin.Context) {
	accs, err := this.pipe.Accounts().Accounts(nil)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(accs, c.Writer)
}

func (this *RestServer) handleAccount(c *gin.Context) {
	addr := c.MustGet("addrBts").([]byte)
	acc, err := this.pipe.Accounts().Account(addr)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(acc, c.Writer)
}

func (this *RestServer) handleStorage(c *gin.Context) {
	addr := c.MustGet("addrBts").([]byte)
	s, err := this.pipe.Accounts().Storage(addr)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(s, c.Writer)
}

func (this *RestServer) handleStorageAt(c *gin.Context) {
	addr := c.MustGet("addrBts").([]byte)
	key := c.MustGet("keyBts").([]byte)
	sa, err := this.pipe.Accounts().StorageAt(addr, key)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(sa, c.Writer)
}

// ********************************* Blockchain *********************************

func (this *RestServer) handleBlockchainInfo(c *gin.Context) {
	bci, err := this.pipe.Blockchain().Info()
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(bci, c.Writer)
}

func (this *RestServer) handleGenesisHash(c *gin.Context) {
	gh, err := this.pipe.Blockchain().GenesisHash()
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(&ep.GenesisHash{gh}, c.Writer)
}

func (this *RestServer) handleChainId(c *gin.Context) {
	cId, err := this.pipe.Blockchain().ChainId()
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(&ep.ChainId{cId}, c.Writer)
}

func (this *RestServer) handleLatestBlockHeight(c *gin.Context) {
	lbh, err := this.pipe.Blockchain().LatestBlockHeight()
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(&ep.LatestBlockHeight{lbh}, c.Writer)
}

func (this *RestServer) handleLatestBlock(c *gin.Context) {
	lb, err := this.pipe.Blockchain().LatestBlock()
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(lb, c.Writer)
}

func (this *RestServer) handleBlocks(c *gin.Context) {
	// TODO fix when query structure has been decided on.
	// rfd := &RangeFilterData{0, 0}
	//blocks, err := this.pipe.Blockchain().Blocks(rfd)
	//if err != nil {
	//	c.AbortWithError(500, err)
	//}
	//c.Writer.WriteHeader(200)
	//this.codec.Encode(blocks, c.Writer)
}

func (this *RestServer) handleBlock(c *gin.Context) {
	height := c.MustGet("height").(uint)
	block, err := this.pipe.Blockchain().Block(height)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(block, c.Writer)
}

// ********************************* Consensus *********************************
func (this *RestServer) handleConsensusState(c *gin.Context) {

	cs, err := this.pipe.Consensus().State()
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(cs, c.Writer)
}

func (this *RestServer) handleValidatorList(c *gin.Context) {
	vl, err := this.pipe.Consensus().Validators()
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(vl, c.Writer)
}

// ********************************* Events *********************************

func (this *RestServer) handleEventSubscribe(c *gin.Context) {
	param := &EventIdParam{}
	errD := this.codec.Decode(param, c.Request.Body)
	if errD != nil {
		c.AbortWithError(500, errD)
	}
	subId, err := this.eventSubs.add(param.EventId)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(&ep.EventSub{subId}, c.Writer)
}

func (this *RestServer) handleEventPoll(c *gin.Context) {
	subId := c.MustGet("id").(string)
	data, err := this.eventSubs.poll(subId)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(&ep.PollResponse{data}, c.Writer)
}

func (this *RestServer) handleEventUnsubscribe(c *gin.Context) {
	subId := c.MustGet("id").(string)
	err := this.eventSubs.remove(subId)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(&ep.EventUnsub{true}, c.Writer)
}

// ********************************* Network *********************************

func (this *RestServer) handleNetworkInfo(c *gin.Context) {
	nInfo, err := this.pipe.Net().Info()
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(nInfo, c.Writer)
}

func (this *RestServer) handleMoniker(c *gin.Context) {
	moniker, err := this.pipe.Net().Moniker()
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(&ep.Moniker{moniker}, c.Writer)
}

func (this *RestServer) handleListening(c *gin.Context) {
	listening, err := this.pipe.Net().Listening()
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(&ep.Listening{listening}, c.Writer)
}

func (this *RestServer) handleListeners(c *gin.Context) {
	listeners, err := this.pipe.Net().Listeners()
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(&ep.Listeners{listeners}, c.Writer)
}

func (this *RestServer) handlePeers(c *gin.Context) {
	peers, err := this.pipe.Net().Peers()
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(peers, c.Writer)
}

func (this *RestServer) handlePeer(c *gin.Context) {
	address := c.MustGet("address").(string)
	peer, err := this.pipe.Net().Peer(address)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(peer, c.Writer)
}

// ********************************* Transactions *********************************

func (this *RestServer) handleBroadcastTx(c *gin.Context) {
	param := &TxParam{}
	errD := this.codec.Decode(param, c.Request.Body)
	if errD != nil {
		c.AbortWithError(500, errD)
	}
	receipt, err := this.pipe.Txs().BroadcastTx(param.Tx)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(receipt, c.Writer)
}

func (this *RestServer) handleUnconfirmedTxs(c *gin.Context) {

	txs, err := this.pipe.Txs().UnconfirmedTxs()
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(txs, c.Writer)
}

func (this *RestServer) handleCall(c *gin.Context) {
	param := &CallParam{}
	errD := this.codec.Decode(param, c.Request.Body)
	if errD != nil {
		c.AbortWithError(500, errD)
	}
	call, err := this.pipe.Txs().Call(param.Address, param.Data)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(call, c.Writer)
}

func (this *RestServer) handleCallCode(c *gin.Context) {
	param := &CallParam{}
	errD := this.codec.Decode(param, c.Request.Body)
	if errD != nil {
		c.AbortWithError(500, errD)
	}
	call, err := this.pipe.Txs().Call(param.Address, param.Data)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(call, c.Writer)
}

func (this *RestServer) handleTransact(c *gin.Context) {
	param := &TransactParam{}
	errD := this.codec.Decode(param, c.Request.Body)
	if errD != nil {
		c.AbortWithError(500, errD)
	}
	receipt, err := this.pipe.Txs().Transact(param.PrivKey, param.Address, param.Data, param.GasLimit, param.Fee)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(receipt, c.Writer)
}

func (this *RestServer) handleSignTx(c *gin.Context) {
	param := &SignTxParam{}
	errD := this.codec.Decode(param, c.Request.Body)
	if errD != nil {
		c.AbortWithError(500, errD)
	}
	tx, err := this.pipe.Txs().SignTx(param.Tx, param.PrivAccounts)
	if err != nil {
		c.AbortWithError(500, err)
	}
	c.Writer.WriteHeader(200)
	this.codec.Encode(tx, c.Writer)
}

// ********************************* Middleware *********************************

func addressParam(c *gin.Context) {
	addr := c.Param("address")
	if !isAddress(addr) {
		c.AbortWithError(400, fmt.Errorf("Malformed address param: "+addr))
	}
	bts, _ := hex.DecodeString(addr)
	c.Set("addrBts", bts)
}

func keyParam(c *gin.Context) {
	key := c.Param("key")
	bts, err := hex.DecodeString(key)
	if err != nil {
		c.AbortWithError(400, err)
	}
	c.Set("keyBts", bts)
}

func heightParam(c *gin.Context) {
	h, err := strconv.Atoi(c.Param("height"))
	if err != nil {
		c.AbortWithError(400, err)
	}
	if h < 0 {
		c.AbortWithError(400, fmt.Errorf("Negative number used as height."))
	}
	c.Set("height", uint(h))
}

func subIdParam(c *gin.Context) {
	subId := c.Param("id")
	if len(subId) != 64 || !isHex(subId) {
		c.AbortWithError(400, fmt.Errorf("Malformed event id"))
	}
	c.Set("id", subId)
}

// TODO
func peerAddressParam(c *gin.Context) {
	subId := c.Param("address")
	c.Set("address", subId)
}


func parseQuery(c *gin.Context){
	q := c.Query("q")
	if q == "" {
		c.Set("filters", nil)
	} else {
		data, err := _parseQuery(q)
		if err != nil {
			c.AbortWithError(400, err)
		}
		c.Set("filters", data)	
	}
}

func _parseQuery(queryString string) ([]*ep.FilterData, error) {
	if queryString == "" {
		return nil, nil
	}
	filters := strings.Split(queryString, "+")
	fdArr := []*ep.FilterData{}
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
func toFilterData(field, stmt string) (*ep.FilterData, *ep.FilterData, error) {
	// In case statement is empty
	if stmt == "" {
		return &ep.FilterData{field, "==", ""}, nil, nil
	}
	// Simple routine based on string splitting. TODO add quoted range query.
	if stmt[0] == '>' || stmt[0] == '<' || stmt[0] == '=' || stmt[0] == '!' {
		// This means a normal operator. If one character then stop, otherwise
		// peek at next and check if it's a "=".
		
		if len(stmt) == 1 {
			return &ep.FilterData{field, stmt[0:1], ""}, nil, nil
		} else if stmt[1] == '=' {
			return &ep.FilterData{field, stmt[:2], stmt[2:]}, nil, nil
		} else {
			return &ep.FilterData{field, stmt[0:1], stmt[1:]}, nil, nil
		}
	} else {
		// Either we have a range query here or a malformed query.
		rng := strings.Split(stmt, "..")
		// This is for when there is no op, but the value is not empty.
		if len(rng) == 1 {
			return &ep.FilterData{field, "==", stmt}, nil, nil
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
		return &ep.FilterData{field, ">=", min}, &ep.FilterData{field, "<=", max}, nil
	}
	return nil, nil, nil
}