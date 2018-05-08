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

package keys

import (
	"context"
	"fmt"
	"time"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/keys/pbkeys"
	"github.com/hyperledger/burrow/logging"
	"google.golang.org/grpc"
)

type KeyClient interface {
	// Sign returns the signature bytes for given hash signed with the key associated with signAddress
	Sign(signAddress crypto.Address, hash []byte) (signature crypto.Signature, err error)

	// PublicKey returns the public key associated with a given address
	PublicKey(address crypto.Address) (publicKey crypto.PublicKey, err error)

	// Generate requests that a key be generate within the keys instance and returns the address
	Generate(keyName string, keyType crypto.CurveType) (keyAddress crypto.Address, err error)

	// Returns nil if the keys instance is healthy, error otherwise
	HealthCheck() error
}

var _ KeyClient = (*localKeyClient)(nil)
var _ KeyClient = (*remoteKeyClient)(nil)

type localKeyClient struct {
	logger *logging.Logger
}

type remoteKeyClient struct {
	rpcAddress string
	kc         pbkeys.KeysClient
	logger     *logging.Logger
}

func (l localKeyClient) Sign(signAddress crypto.Address, hash []byte) (signature crypto.Signature, err error) {
	resp, err := GlobalKeyServer.Sign(nil, &pbkeys.SignRequest{Address: signAddress.String(), Hash: hash})
	if err != nil {
		return crypto.Signature{}, err
	}
	curveType, err := crypto.CurveTypeFromString(resp.GetCurvetype())
	if err != nil {
		return crypto.Signature{}, err
	}
	return crypto.SignatureFromBytes(resp.GetSignature(), curveType)
}

func (l localKeyClient) PublicKey(address crypto.Address) (publicKey crypto.PublicKey, err error) {
	resp, err := GlobalKeyServer.PublicKey(nil, &pbkeys.PubRequest{Address: address.String()})
	if err != nil {
		return crypto.PublicKey{}, err
	}
	curveType, err := crypto.CurveTypeFromString(resp.GetCurvetype())
	if err != nil {
		return crypto.PublicKey{}, err
	}
	return crypto.PublicKeyFromBytes(resp.GetPub(), curveType)
}

// Generate requests that a key be generate within the keys instance and returns the address
func (l localKeyClient) Generate(keyName string, curveType crypto.CurveType) (keyAddress crypto.Address, err error) {
	resp, err := GlobalKeyServer.GenerateKey(nil, &pbkeys.GenRequest{Keyname: keyName, Curvetype: curveType.String()})
	if err != nil {
		return crypto.Address{}, err
	}
	return crypto.AddressFromHexString(resp.GetAddress())
}

// Returns nil if the keys instance is healthy, error otherwise
func (l localKeyClient) HealthCheck() error {
	return nil
}

func (l remoteKeyClient) Sign(signAddress crypto.Address, message []byte) (signature crypto.Signature, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req := pbkeys.SignRequest{Address: signAddress.String(), Hash: message}
	l.logger.TraceMsg("Sending Sign request to remote key server: ", fmt.Sprintf("%v", req))
	resp, err := l.kc.Sign(ctx, &req)
	if err != nil {
		l.logger.TraceMsg("Received Sign request error response: ", err)
		return crypto.Signature{}, err
	}
	l.logger.TraceMsg("Received Sign response to remote key server: %v", resp)
	curveType, err := crypto.CurveTypeFromString(resp.GetCurvetype())
	if err != nil {
		return crypto.Signature{}, err
	}
	return crypto.SignatureFromBytes(resp.GetSignature(), curveType)
}

func (l remoteKeyClient) PublicKey(address crypto.Address) (publicKey crypto.PublicKey, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req := pbkeys.PubRequest{Address: address.String()}
	l.logger.TraceMsg("Sending PublicKey request to remote key server: ", fmt.Sprintf("%v", req))
	resp, err := l.kc.PublicKey(ctx, &req)
	if err != nil {
		l.logger.TraceMsg("Received PublicKey error response: ", err)
		return crypto.PublicKey{}, err
	}
	curveType, err := crypto.CurveTypeFromString(resp.GetCurvetype())
	if err != nil {
		return crypto.PublicKey{}, err
	}
	l.logger.TraceMsg("Received PublicKey response to remote key server: ", fmt.Sprintf("%v", resp))
	return crypto.PublicKeyFromBytes(resp.GetPub(), curveType)
}

// Generate requests that a key be generate within the keys instance and returns the address
func (l remoteKeyClient) Generate(keyName string, curveType crypto.CurveType) (keyAddress crypto.Address, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req := pbkeys.GenRequest{Keyname: keyName, Curvetype: curveType.String()}
	l.logger.TraceMsg("Sending Generate request to remote key server: ", fmt.Sprintf("%v", req))
	resp, err := l.kc.GenerateKey(ctx, &req)
	if err != nil {
		l.logger.TraceMsg("Received Generate error response: ", err)
		return crypto.Address{}, err
	}
	l.logger.TraceMsg("Received Generate response to remote key server: ", fmt.Sprintf("%v", resp))
	return crypto.AddressFromHexString(resp.GetAddress())
}

// Returns nil if the keys instance is healthy, error otherwise
func (l remoteKeyClient) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := l.kc.List(ctx, &pbkeys.Name{""})
	return err
}

// keyClient.New returns a new monax-keys client for provided rpc location
// Monax-keys connects over http request-responses
func NewKeyClient(rpcAddress string, logger *logging.Logger) KeyClient {
	logger = logger.WithScope("NewKeyClient")
	var client KeyClient
	if rpcAddress != "" {
		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())
		conn, err := grpc.Dial(rpcAddress, opts...)
		if err != nil {
			// FIXME: we should return error, or handle this when the
			// key client is used for the first time
			panic("Failed to connect to grpc server")
		}
		kc := pbkeys.NewKeysClient(conn)

		client = remoteKeyClient{kc: kc, rpcAddress: rpcAddress, logger: logger}
	} else {
		ks, err := newKeyStore()
		if err != nil {
			panic("Failed to start keys store")
		}

		GlobalKeystore = ks

		client = localKeyClient{logger: logger}
	}
	return client
}

type signer struct {
	keyClient KeyClient
	address   crypto.Address
}

func (ms *signer) Sign(messsage []byte) (crypto.Signature, error) {
	signature, err := ms.keyClient.Sign(ms.address, messsage)
	if err != nil {
		return crypto.Signature{}, err
	}
	return signature, nil
}

// Creates a Signer that assumes the address holds an Ed25519 key
func Signer(keyClient KeyClient, address crypto.Address) crypto.Signer {
	// TODO: we can do better than this and return a typed signature when we reform the keys service
	return &signer{
		keyClient: keyClient,
		address:   address,
	}
}

type keyAddressable struct {
	publicKey crypto.PublicKey
	address   crypto.Address
}

func (ka *keyAddressable) Address() crypto.Address {
	return ka.address
}

func (ka *keyAddressable) PublicKey() crypto.PublicKey {
	return ka.publicKey
}

func Addressable(keyClient KeyClient, address crypto.Address) (acm.Addressable, error) {
	pubKey, err := keyClient.PublicKey(address)
	if err != nil {
		return nil, err
	}
	return &keyAddressable{
		address:   address,
		publicKey: pubKey,
	}, nil
}
