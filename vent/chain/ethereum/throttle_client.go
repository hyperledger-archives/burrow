package ethereum

import (
	"time"

	"github.com/hyperledger/burrow/logging"

	"github.com/hyperledger/burrow/rpc/web3/ethclient"
)

type ThrottleClient interface {
	EthClient
	Throttle()
}

type throttleClient struct {
	*Throttler
	client EthClient
}

func NewThrottleClient(client EthClient, maxRequests int, timeBase time.Duration, logger *logging.Logger) *throttleClient {
	return &throttleClient{
		Throttler: NewThrottler(maxRequests, timeBase, timeBase,
			logger.WithScope("ThrottleClient").
				With("max_requests", maxRequests, "time_base", timeBase.String())),
		client: client,
	}
}

func (t *throttleClient) GetLogs(filter *ethclient.Filter) ([]*ethclient.EthLog, error) {
	t.addNow()
	return t.client.GetLogs(filter)
}

func (t *throttleClient) BlockNumber() (uint64, error) {
	t.addNow()
	return t.client.BlockNumber()
}

func (t *throttleClient) GetBlockByNumber(height string) (*ethclient.Block, error) {
	t.addNow()
	return t.client.GetBlockByNumber(height)
}

func (t *throttleClient) NetVersion() (string, error) {
	t.addNow()
	return t.client.NetVersion()
}

func (t *throttleClient) Web3ClientVersion() (string, error) {
	t.addNow()
	return t.client.Web3ClientVersion()
}

func (t *throttleClient) Syncing() (bool, error) {
	t.addNow()
	return t.client.Syncing()
}
