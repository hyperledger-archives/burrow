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

package tm

import (
	"fmt"

	"github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/blockchain"
	core_types "github.com/hyperledger/burrow/core/types"
	"github.com/hyperledger/burrow/event"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/logging"
	logging_types "github.com/hyperledger/burrow/logging/types"
	"github.com/hyperledger/burrow/rpc/tm/types"
	"github.com/hyperledger/burrow/txs"
	"github.com/hyperledger/sawtooth-core/families/seth/src/burrow/word256"
	abci_types "github.com/tendermint/abci/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/rpc/lib/types"
	tm_types "github.com/tendermint/tendermint/types"
)

type service struct {
	logger         logging_types.InfoTraceLogger
	eventEmitter   event.EventEmitter
	blockchain     blockchain.BlockchainTip
	blockStore     tm_types.BlockStoreRPC
	tendermintNode *node.Node
}

func (s *service) Subscribe(wsCtx rpctypes.WSRPCContext,
	eventId string) (*types.ResultSubscribe, error) {
	// NOTE: RPCResponses of subscribed events have id suffix "#event"
	// TODO: we really ought to allow multiple subscriptions from the same client address
	// to the same event. The code as it stands reflects the somewhat broken tendermint
	// implementation. We can use GenerateSubId to randomize the subscriptions id
	// and return it in the result. This would require clients to hang on to a
	// subscription id if they wish to unsubscribe, but then again they can just
	// drop their connection
	subscriptionId, err := event.GenerateSubId()
	if err != nil {
		return nil, err
		logging.InfoMsg(s.logger, "Subscribing to event",
			"eventId", eventId, "subscriptionId", subscriptionId)
	}
	s.eventEmitter.Subscribe(subscriptionId, eventId,
		func(eventData evm.EventData) {
			result := types.BurrowResult(
				&types.ResultEvent{
					Event: eventId,
					Data:  evm.EventData(eventData)})
			// NOTE: EventSwitch callbacks must be nonblocking
			wsCtx.GetRemoteAddr()
			// NOTE: EventSwitch callbacks must be nonblocking
			wsCtx.TryWriteRPCResponse(
				rpctypes.NewRPCResponse(wsCtx.Request.ID+"#event", &result, ""))
		})
	return &types.ResultSubscribe{
		SubscriptionId: subscriptionId,
		Event:          eventId,
	}, nil
}

func (s *service) Unsubscribe(wsCtx rpctypes.WSRPCContext,
	subscriptionId string) (*types.ResultUnsubscribe, error) {
	err := s.eventEmitter.Unsubscribe(subscriptionId)
	if err != nil {
		return nil, err
	} else {
		return &types.ResultUnsubscribe{SubscriptionId: subscriptionId}, nil
	}
}

func (s *service) Status() (*types.ResultStatus, error) {
	latestHeight := s.blockchain.LastBlockHeight()
	var (
		latestBlockMeta *tm_types.BlockMeta
		latestBlockHash []byte
		latestBlockTime int64
	)
	if latestHeight != 0 {
		latestBlockMeta = s.blockStore.LoadBlockMeta(int(latestHeight))
		latestBlockHash = latestBlockMeta.Header.Hash()
		latestBlockTime = latestBlockMeta.Header.Time.UnixNano()
	}
	return &types.ResultStatus{
		NodeInfo:          s.tendermintNode.NodeInfo(),
		GenesisHash:       s.tendermintNode.GenesisDoc().AppHash,
		PubKey:            s.tendermintNode.PrivValidator().PubKey,
		LatestBlockHash:   latestBlockHash,
		LatestBlockHeight: latestHeight,
		LatestBlockTime:   latestBlockTime}, nil
}

func (s *service) ChainId() (*types.ResultChainId, error) {
	if s.blockchain == nil {
		return nil, fmt.Errorf("Blockchain not initialised in burrowmint s.")
	}
	chainId := s.blockchain.ChainId()

	return &types.ResultChainId{
		ChainName:   chainId, // MARMOT: copy ChainId for ChainName as a placehodlder
		ChainId:     chainId,
		GenesisHash: s.GenesisHash(),
	}, nil
}

func (s *service) NetInfo() (*types.ResultNetInfo, error) {
	listening := s.consensusEngine.IsListening()
	listeners := []string{}
	for _, listener := range s.consensusEngine.Listeners() {
		listeners = append(listeners, listener.String())
	}
	peers := s.consensusEngine.Peers()
	return &types.ResultNetInfo{
		Listening: listening,
		Listeners: listeners,
		Peers:     peers,
	}, nil
}

func (s *service) Genesis() (*types.ResultGenesis, error) {
	return &types.ResultGenesis{
		// TODO: [ben] sharing pointer to unmutated GenesisDoc, but is not immutable
		Genesis: s.genesisDoc,
	}, nil
}

// Accounts
func (s *service) GetAccount(address []byte) (*types.ResultGetAccount,
	error) {
	cache := s.burrowMint.GetCheckCache()
	account := cache.GetAccount(address)
	return &types.ResultGetAccount{Account: account}, nil
}

func (s *service) ListAccounts() (*types.ResultListAccounts, error) {
	var blockHeight int
	var accounts []*account.ConcreteAccount
	state := s.burrowMint.GetState()
	blockHeight = state.LastBlockHeight
	state.GetAccounts().Iterate(func(key []byte, value []byte) bool {
		accounts = append(accounts, account.DecodeAccount(value))
		return false
	})
	return &types.ResultListAccounts{blockHeight, accounts}, nil
}

func (s *service) GetStorage(address, key []byte) (*types.ResultGetStorage,
	error) {
	state := s.burrowMint.GetState()
	// state := consensusState.GetState()
	account := state.GetAccount(address)
	if account == nil {
		return nil, fmt.Errorf("UnknownAddress: %X", address)
	}
	storageRoot := account.StorageRoot
	storageTree := state.LoadStorage(storageRoot)

	_, value, exists := storageTree.Get(
		word256.LeftPadWord256(key).Bytes())
	if !exists {
		// value == nil {
		return &types.ResultGetStorage{key, nil}, nil
	}
	return &types.ResultGetStorage{key, value}, nil
}

func (s *service) DumpStorage(address []byte) (*types.ResultDumpStorage,
	error) {
	state := s.burrowMint.GetState()
	account := state.GetAccount(address)
	if account == nil {
		return nil, fmt.Errorf("UnknownAddress: %X", address)
	}
	storageRoot := account.StorageRoot
	storageTree := state.LoadStorage(storageRoot)
	storageItems := []types.StorageItem{}
	storageTree.Iterate(func(key []byte, value []byte) bool {
		storageItems = append(storageItems, types.StorageItem{key,
			value})
		return false
	})
	return &types.ResultDumpStorage{storageRoot, storageItems}, nil
}

// Call
// NOTE: this function is used from 46657 and has sibling on 1337
// in transactor.go
// TODO: [ben] resolve incompatibilities in byte representation for 0.12.0 release
func (s *service) Call(fromAddress, toAddress, data []byte) (*types.ResultCall,
	error) {
	if vm.RegisteredNativeContract(word256.LeftPadWord256(toAddress)) {
		return nil, fmt.Errorf("Attempt to call native contract at address "+
			"%X, but native contracts can not be called directly. Use a deployed "+
			"contract that calls the native function instead.", toAddress)
	}
	st := s.burrowMint.GetState()
	cache := state.NewBlockCache(st)
	outAcc := cache.GetAccount(toAddress)
	if outAcc == nil {
		return nil, fmt.Errorf("Account %x does not exist", toAddress)
	}
	if fromAddress == nil {
		fromAddress = []byte{}
	}
	callee := toVMAccount(outAcc)
	caller := &vm.Account{Address: word256.LeftPadWord256(fromAddress)}
	txCache := state.NewTxCache(cache)
	gasLimit := st.GetGasLimit()
	params := vm.Params{
		BlockHeight: int64(st.LastBlockHeight),
		BlockHash:   word256.LeftPadWord256(st.LastBlockHash),
		BlockTime:   st.LastBlockTime.Unix(),
		GasLimit:    gasLimit,
	}

	vmach := vm.NewVM(txCache, vm.DefaultDynamicMemoryProvider, params,
		caller.Address, nil)
	gas := gasLimit
	ret, err := vmach.Call(caller, callee, callee.Code, data, 0, &gas)
	if err != nil {
		return nil, err
	}
	gasUsed := gasLimit - gas
	// here return bytes are not hex encoded; on the sibling function
	// they are
	return &types.ResultCall{Return: ret, GasUsed: gasUsed}, nil
}

func (s *service) CallCode(fromAddress, code, data []byte) (*types.ResultCall,
	error) {
	st := s.burrowMint.GetState()
	cache := s.burrowMint.GetCheckCache()
	callee := &vm.Account{Address: word256.LeftPadWord256(fromAddress)}
	caller := &vm.Account{Address: word256.LeftPadWord256(fromAddress)}
	txCache := state.NewTxCache(cache)
	gasLimit := st.GetGasLimit()
	params := vm.Params{
		BlockHeight: int64(st.LastBlockHeight),
		BlockHash:   word256.LeftPadWord256(st.LastBlockHash),
		BlockTime:   st.LastBlockTime.Unix(),
		GasLimit:    gasLimit,
	}

	vmach := vm.NewVM(txCache, vm.DefaultDynamicMemoryProvider, params,
		caller.Address, nil)
	gas := gasLimit
	ret, err := vmach.Call(caller, callee, code, data, 0, &gas)
	if err != nil {
		return nil, err
	}
	gasUsed := gasLimit - gas
	return &types.ResultCall{Return: ret, GasUsed: gasUsed}, nil
}

// TODO: [ben] deprecate as we should not allow unsafe behaviour
// where a user is allowed to send a private key over the wire,
// especially unencrypted.
func (s *service) SignTransaction(tx txs.Tx,
	privAccounts []*account.ConcretePrivateAccount) (*types.ResultSignTx,
	error) {

	for i, privAccount := range privAccounts {
		if privAccount == nil || privAccount.PrivKey == nil {
			return nil, fmt.Errorf("Invalid (empty) privAccount @%v", i)
		}
	}
	switch tx.(type) {
	case *txs.SendTx:
		sendTx := tx.(*txs.SendTx)
		for i, input := range sendTx.Inputs {
			input.PubKey = privAccounts[i].PubKey
			input.Signature = privAccounts[i].Sign(s.transactor.chainID, sendTx)
		}
	case *txs.CallTx:
		callTx := tx.(*txs.CallTx)
		callTx.Input.PubKey = privAccounts[0].PubKey
		callTx.Input.Signature = privAccounts[0].Sign(s.transactor.chainID, callTx)
	case *txs.BondTx:
		bondTx := tx.(*txs.BondTx)
		// the first privaccount corresponds to the BondTx pub key.
		// the rest to the inputs
		bondTx.Signature = privAccounts[0].Sign(s.transactor.chainID, bondTx).(crypto.SignatureEd25519)
		for i, input := range bondTx.Inputs {
			input.PubKey = privAccounts[i+1].PubKey
			input.Signature = privAccounts[i+1].Sign(s.transactor.chainID, bondTx)
		}
	case *txs.UnbondTx:
		unbondTx := tx.(*txs.UnbondTx)
		unbondTx.Signature = privAccounts[0].Sign(s.transactor.chainID, unbondTx).(crypto.SignatureEd25519)
	case *txs.RebondTx:
		rebondTx := tx.(*txs.RebondTx)
		rebondTx.Signature = privAccounts[0].Sign(s.transactor.chainID, rebondTx).(crypto.SignatureEd25519)
	}
	return &types.ResultSignTx{tx}, nil
}

// Name registry
func (s *service) GetName(name string) (*types.ResultGetName, error) {
	currentState := s.burrowMint.GetState()
	entry := currentState.GetNameRegEntry(name)
	if entry == nil {
		return nil, fmt.Errorf("Name %s not found", name)
	}
	return &types.ResultGetName{entry}, nil
}

func (s *service) ListNames() (*types.ResultListNames, error) {
	var blockHeight int
	var names []*core_types.NameRegEntry
	currentState := s.burrowMint.GetState()
	blockHeight = currentState.LastBlockHeight
	currentState.GetNames().Iterate(func(key []byte, value []byte) bool {
		names = append(names, state.DecodeNameRegEntry(value))
		return false
	})
	return &types.ResultListNames{blockHeight, names}, nil
}

func (s *service) broadcastTx(tx txs.Tx,
	callback func(res *abci_types.Response)) (*types.ResultBroadcastTx, error) {

	txBytes, err := txs.EncodeTx(tx)
	if err != nil {
		return nil, fmt.Errorf("Error encoding transaction: %v", err)
	}
	err = s.consensusEngine.BroadcastTransaction(txBytes, callback)
	if err != nil {
		return nil, fmt.Errorf("Error broadcasting transaction: %v", err)
	}
	return &types.ResultBroadcastTx{}, nil
}

// Memory pool
// NOTE: txs must be signed
func (s *service) BroadcastTxAsync(tx txs.Tx) (*types.ResultBroadcastTx, error) {
	return s.broadcastTx(tx, nil)
}

func (s *service) BroadcastTxSync(tx txs.Tx) (*types.ResultBroadcastTx, error) {
	responseChannel := make(chan *abci_types.Response, 1)
	_, err := s.broadcastTx(tx,
		func(res *abci_types.Response) {
			responseChannel <- res
		})
	if err != nil {
		return nil, err
	}
	// NOTE: [ben] This Response is set in /consensus/tendermint/local_client.go
	// a call to Application, here implemented by BurrowMint, over local callback,
	// or abci RPC call.  Hence the result is determined by BurrowMint/burrowmint.go
	// CheckTx() Result (Result converted to ReqRes into Response returned here)
	// NOTE: [ben] BroadcastTx just calls CheckTx in Tendermint (oddly... [Silas])
	response := <-responseChannel
	responseCheckTx := response.GetCheckTx()
	if responseCheckTx == nil {
		return nil, fmt.Errorf("Error, application did not return CheckTx response.")
	}
	resultBroadCastTx := &types.ResultBroadcastTx{
		Code: responseCheckTx.Code,
		Data: responseCheckTx.Data,
		Log:  responseCheckTx.Log,
	}
	switch responseCheckTx.Code {
	case abci_types.CodeType_OK:
		return resultBroadCastTx, nil
	case abci_types.CodeType_EncodingError:
		return resultBroadCastTx, fmt.Errorf(resultBroadCastTx.Log)
	case abci_types.CodeType_InternalError:
		return resultBroadCastTx, fmt.Errorf(resultBroadCastTx.Log)
	default:
		logging.InfoMsg(s.logger, "Unknown error returned from Tendermint CheckTx on BroadcastTxSync",
			"application", "burrowmint",
			"abci_code_type", responseCheckTx.Code,
			"abci_log", responseCheckTx.Log,
		)
		return resultBroadCastTx, fmt.Errorf("Unknown error returned: " + responseCheckTx.Log)
	}
}

func (s *service) ListUnconfirmedTxs(maxTxs int) (*types.ResultListUnconfirmedTxs, error) {
	// Get all transactions for now
	transactions, err := s.consensusEngine.ListUnconfirmedTxs(maxTxs)
	if err != nil {
		return nil, err
	}
	return &types.ResultListUnconfirmedTxs{
		N:   len(transactions),
		Txs: transactions,
	}, nil
}

// Returns the current blockchain height and metadata for a range of blocks
// between minHeight and maxHeight. Only returns maxBlockLookback block metadata
// from the top of the range of blocks.
// Passing 0 for maxHeight sets the upper height of the range to the current
// blockchain height.
func (s *service) BlockchainInfo(minHeight, maxHeight,
	maxBlockLookback int) (*types.ResultBlockchainInfo, error) {

	latestHeight := s.blockchain.Height()

	if maxHeight < 1 {
		maxHeight = latestHeight
	} else {
		maxHeight = imath.MinInt(latestHeight, maxHeight)
	}
	if minHeight < 1 {
		minHeight = imath.MaxInt(1, maxHeight-maxBlockLookback)
	}

	blockMetas := []*tm_types.BlockMeta{}
	for height := maxHeight; height >= minHeight; height-- {
		blockMeta := s.blockchain.BlockMeta(height)
		blockMetas = append(blockMetas, blockMeta)
	}

	return &types.ResultBlockchainInfo{
		LastHeight: latestHeight,
		BlockMetas: blockMetas,
	}, nil
}

func (s *service) GetBlock(height int) (*types.ResultGetBlock, error) {
	return &types.ResultGetBlock{
		Block:     s.blockchain.Block(height),
		BlockMeta: s.blockchain.BlockMeta(height),
	}, nil
}

func (s *service) ListValidators() (*types.ResultListValidators, error) {
	validators := s.consensusEngine.ListValidators()
	consensusState := s.consensusEngine.ConsensusState()
	// TODO: when we reintroduce support for bonding and unbonding update this
	// to reflect the mutable bonding state
	return &types.ResultListValidators{
		BlockHeight:         consensusState.Height,
		BondedValidators:    validators,
		UnbondingValidators: nil,
	}, nil
}

func (s *service) DumpConsensusState() (*types.ResultDumpConsensusState, error) {
	statesMap := s.consensusEngine.PeerConsensusStates()
	peerStates := make([]*types.ResultPeerConsensusState, len(statesMap))
	for key, peerState := range statesMap {
		peerStates = append(peerStates, &types.ResultPeerConsensusState{
			PeerKey:            key,
			PeerConsensusState: peerState,
		})
	}
	dump := types.ResultDumpConsensusState{
		ConsensusState:      s.consensusEngine.ConsensusState(),
		PeerConsensusStates: peerStates,
	}
	return &dump, nil
}
