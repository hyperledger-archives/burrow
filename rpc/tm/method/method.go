package method

const (
	Subscribe   = "subscribe"
	Unsubscribe = "unsubscribe"

	// Status
	Status  = "status"
	NetInfo = "net_info"

	// Accounts
	ListAccounts = "list_accounts"
	GetAccount   = "get_account"
	GetStorage   = "get_storage"
	DumpStorage  = "dump_storage"

	// Simulated call
	Call     = "call"
	CallCode = "call_code"

	// Names
	GetName     = "get_name"
	ListNames   = "list_names"
	BroadcastTx = "broadcast_tx"

	// Blockchain
	Genesis    = "genesis"
	ChainID    = "chain_id"
	Blockchain = "blockchain"
	GetBlock   = "get_block"

	// Consensus
	ListUnconfirmedTxs = "list_unconfirmed_txs"
	ListValidators     = "list_validators"
	DumpConsensusState = "dump_consensus_state"

	// Private keys and signing
	GeneratePrivateAccount = "unsafe/gen_priv_account"
	SignTx                 = "unsafe/sign_tx"
)
