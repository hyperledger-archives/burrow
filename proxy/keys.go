package proxy

import (
	"context"
	"crypto/sha256"
	"fmt"
	"hash"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/keys"
	hex "github.com/tmthrgd/go-hex"
	"golang.org/x/crypto/ripemd160"
	"google.golang.org/grpc"
)

//------------------------------------------------------------------------
// handlers

type KeysService struct {
	KeyStore *keys.KeyStore
}

func RegisterKeysService(server *grpc.Server, keyStore *keys.KeyStore) {
	keys.RegisterKeysServer(server, &KeysService{KeyStore: keyStore})
}

func (k *KeysService) GenerateKey(ctx context.Context, in *keys.GenRequest) (*keys.GenResponse, error) {
	curveT, err := crypto.CurveTypeFromString(in.CurveType)
	if err != nil {
		return nil, err
	}

	key, err := k.KeyStore.Gen(in.Passphrase, curveT)
	if err != nil {
		return nil, fmt.Errorf("error generating key %s %s", curveT, err)
	}

	if in.KeyName != "" {
		err = k.KeyStore.AddName(in.KeyName, key.Address)
		if err != nil {
			return nil, err
		}
	}

	return &keys.GenResponse{Address: key.Address}, nil
}

func (k *KeysService) PublicKey(ctx context.Context, in *keys.PubRequest) (*keys.PubResponse, error) {
	var addr crypto.Address
	var err error
	if in.Address != nil {
		addr = *in.Address
	} else {
		addr, err = k.KeyStore.GetName(in.GetName())
		if err != nil {
			return nil, err
		}
	}

	// No phrase needed for public key. I hope.
	key, err := k.KeyStore.GetKey("", addr)
	if key == nil {
		return nil, err
	}

	return &keys.PubResponse{CurveType: key.CurveType.String(), PublicKey: key.GetPublicKey().PublicKey}, nil
}

func (k *KeysService) Verify(ctx context.Context, in *keys.VerifyRequest) (*keys.VerifyResponse, error) {
	if in.GetPublicKey() == nil {
		return nil, fmt.Errorf("must provide a pubkey")
	}
	if in.GetMessage() == nil {
		return nil, fmt.Errorf("must provide a message")
	}

	sig := in.Signature
	pubkey, err := crypto.PublicKeyFromBytes(in.GetPublicKey(), sig.GetCurveType())
	if err != nil {
		return nil, err
	}
	err = pubkey.Verify(in.GetMessage(), &sig)
	if err != nil {
		return nil, err
	}

	return &keys.VerifyResponse{}, nil
}

func (k *KeysService) Hash(ctx context.Context, in *keys.HashRequest) (*keys.HashResponse, error) {
	var hasher hash.Hash
	switch in.GetHashtype() {
	case "ripemd160":
		hasher = ripemd160.New()
	case "sha256":
		hasher = sha256.New()
	// case "sha3":
	default:
		return nil, fmt.Errorf("unknown hash type %v", in.GetHashtype())
	}

	hasher.Write(in.GetMessage())

	return &keys.HashResponse{Hash: hex.EncodeUpperToString(hasher.Sum(nil))}, nil
}

func (k *KeysService) ImportJSON(ctx context.Context, in *keys.ImportJSONRequest) (*keys.ImportResponse, error) {
	addr, err := k.KeyStore.ImportJSON(in.GetPassphrase(), in.GetJSON())
	return &keys.ImportResponse{Address: *addr}, err
}

func (k *KeysService) Import(ctx context.Context, in *keys.ImportRequest) (*keys.ImportResponse, error) {
	curveT, err := crypto.CurveTypeFromString(in.GetCurveType())
	if err != nil {
		return nil, err
	}
	key, err := keys.NewKeyFromPriv(curveT, in.GetKeyBytes())
	if err != nil {
		return nil, err
	}

	// store the new key
	if err = k.KeyStore.StoreKey(in.GetPassphrase(), key); err != nil {
		return nil, err
	}

	if in.GetName() != "" {
		err = k.KeyStore.AddName(in.GetName(), key.Address)
		if err != nil {
			return nil, err
		}
	}
	return &keys.ImportResponse{Address: key.Address}, nil
}

func (k *KeysService) List(ctx context.Context, in *keys.ListRequest) (*keys.ListResponse, error) {
	list, err := k.KeyStore.List(in.GetKeyName())
	return &keys.ListResponse{Key: list}, err
}

func (k *KeysService) RemoveName(ctx context.Context, in *keys.RemoveNameRequest) (*keys.RemoveNameResponse, error) {
	if in.GetKeyName() == "" {
		return nil, fmt.Errorf("please specify a name")
	}

	return &keys.RemoveNameResponse{}, k.KeyStore.RmName(in.GetKeyName())
}

func (k *KeysService) AddName(ctx context.Context, in *keys.AddNameRequest) (*keys.AddNameResponse, error) {
	if in.GetKeyname() == "" {
		return nil, fmt.Errorf("please specify a name")
	}

	return &keys.AddNameResponse{}, k.KeyStore.AddName(in.GetKeyname(), in.Address)
}
