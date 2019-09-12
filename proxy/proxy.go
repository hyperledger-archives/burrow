package proxy

import (
	"context"
	"sync"
	"time"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/keys"
	"github.com/hyperledger/burrow/rpc/rpcevents"
	"github.com/hyperledger/burrow/rpc/rpcquery"
	"github.com/hyperledger/burrow/rpc/rpctransact"
	"google.golang.org/grpc"
)

type Proxy struct {
	chainID           string
	lastUpdateChainID time.Time
	query             rpcquery.QueryClient
	events            rpcevents.ExecutionEventsClient
	transact          rpctransact.TransactClient
	keys              KeysService
	cachePeriod       time.Duration
	sequenceCacheLock sync.Mutex
	sequenceCache     map[crypto.Address]sequence
}

type sequence struct {
	sequence   uint64
	lastUpdate time.Time
}

func New(grpcconnection *grpc.ClientConn, server *grpc.Server, keysDir string, AllowBadFilePermissions bool) {
	ks := keys.NewKeyStore(keysDir, AllowBadFilePermissions)

	p := Proxy{
		events:   rpcevents.NewExecutionEventsClient(grpcconnection),
		query:    rpcquery.NewQueryClient(grpcconnection),
		transact: rpctransact.NewTransactClient(grpcconnection),
		keys:     KeysService{KeyStore: ks},
	}

	rpcquery.RegisterQueryServer(server, &p)
	rpcevents.RegisterExecutionEventsServer(server, &p)
	rpctransact.RegisterTransactServer(server, &p)
	keys.RegisterKeysServer(server, &p.keys)
}

func (p *Proxy) BroadcastTxStream(param *rpctransact.TxEnvelopeParam, stream rpctransact.Transact_BroadcastTxStreamServer) error {
	ctx := context.Background()

	// get chain
	if p.chainID == "" || time.Since(p.lastUpdateChainID) > p.cachePeriod {
		status, err := p.query.Status(context.Background(), &rpcquery.StatusParam{})
		if err != nil {
			// log this
			return nil
		}

		p.chainID = status.ChainID
		p.lastUpdateChainID = time.Now()
	}

	txEnv := param.GetEnvelope(p.chainID)

	p.sequenceCacheLock.Lock()
	locked := false
	defer func() {
		if locked {
			p.sequenceCacheLock.Unlock()
		}
	}()

	if len(txEnv.Signatories) == 0 {
		inputs := txEnv.Tx.GetInputs()
		signers := make([]acm.AddressableSigner, len(inputs))

		// Get sequence number for account
		for i, input := range inputs {
			seq, ok := p.sequenceCache[input.Address]
			if !ok || time.Since(seq.lastUpdate) > p.cachePeriod {
				acc, err := p.query.GetAccount(ctx, &rpcquery.GetAccountParam{Address: input.Address})
				if err != nil {
					// FIXME: log this
					return err
				}
				seq = sequence{sequence: acc.GetSequence(), lastUpdate: time.Now()}
			}
			seq.sequence++
			p.sequenceCache[input.Address] = seq
			input.Sequence = seq.sequence
			var err error
			signers[i], err = p.keys.KeyStore.GetKey("", input.Address)
			if err != nil {
				return err
			}
		}

		// sign stuf
		txEnv.Tx.Rehash()

		err := txEnv.Sign(signers...)
		if err != nil {
			return err
		}
	}

	client, err := p.transact.BroadcastTxStream(context.Background(), param)
	if err != nil {
		return err
	}

	for {
		acc, err := client.Recv()
		if err != nil {
			return err
		}
		if acc.Receipt != nil {
			p.sequenceCacheLock.Unlock()
			locked = false
		}
		err = stream.Send(acc)
		if err != nil {
			return err
		}
	}
}
