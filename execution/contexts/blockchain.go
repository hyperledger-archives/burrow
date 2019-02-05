package contexts

import "time"

// Execution's sufficient view of blockchain
type Blockchain interface {
	BlockHash(height uint64) []byte
	LastBlockTime() time.Time
	BlockchainHeight
}

type BlockchainHeight interface {
	LastBlockHeight() uint64
}
