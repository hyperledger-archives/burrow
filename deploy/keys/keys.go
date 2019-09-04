package keys

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/logging"
)

type LocalKeyClient struct {
	crypto.KeyClient
}

var keysTimeout = 5 * time.Second

// Returns an initialized key client to a docker container
// running the keys server
// Adding the Ip address is optional and should only be used
// for passing data
func InitKeyClient(keysUrl string) (*LocalKeyClient, error) {
	aliveCh := make(chan struct{})
	localKeyClient, err := keys.NewRemoteKeyClient(keysUrl, logging.NewNoopLogger())
	if err != nil {
		return nil, err
	}

	err = localKeyClient.HealthCheck()

	go func() {
		for err != nil {
			err = localKeyClient.HealthCheck()
		}
		aliveCh <- struct{}{}
	}()
	select {
	case <-time.After(keysTimeout):
		return nil, fmt.Errorf("keys instance did not become responsive after %s: %v", keysTimeout, err)
	case <-aliveCh:
		return &LocalKeyClient{localKeyClient}, nil
	}
}

// Keyclient returns a list of keys that it is aware of.
// params:
// host - search for keys on the host
// container - search for keys on the container
// quiet - don't print output, just return the list you find
func (keys *LocalKeyClient) ListKeys(keysPath string, quiet bool, logger *logging.Logger) ([]string, error) {
	var result []string
	addrs, err := ioutil.ReadDir(keysPath)
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		result = append(result, addr.Name())
	}
	if !quiet {
		if len(addrs) == 0 {
			logger.InfoMsg("No keys found on host")
		} else {
			// First key.
			logger.InfoMsg("The keys on host", result)
		}
	}

	return result, nil
}
