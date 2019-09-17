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
)

//------------------------------------------------------------------------
// handlers

func (p *Proxy) GenerateKey(ctx context.Context, in *keys.GenRequest) (*keys.GenResponse, error) {
	curveT, err := crypto.CurveTypeFromString(in.CurveType)
	if err != nil {
		return nil, err
	}

	key, err := p.keys.Gen(in.Passphrase, curveT)
	if err != nil {
		return nil, fmt.Errorf("error generating key %s %s", curveT, err)
	}

	if in.KeyName != "" {
		err = p.keys.AddName(in.KeyName, key.Address)
		if err != nil {
			return nil, err
		}
	}

	return &keys.GenResponse{Address: key.Address}, nil
}

func (p *Proxy) PublicKey(ctx context.Context, in *keys.PubRequest) (*keys.PubResponse, error) {
	var addr crypto.Address
	var err error
	if in.Address != nil {
		addr = *in.Address
	} else {
		addr, err = p.keys.GetName(in.GetName())
		if err != nil {
			return nil, err
		}
	}

	// No phrase needed for public key. I hope.
	key, err := p.keys.GetKey("", addr)
	if key == nil {
		return nil, err
	}

	return &keys.PubResponse{CurveType: key.CurveType.String(), PublicKey: key.GetPublicKey().PublicKey}, nil
}

func (p *Proxy) Sign(ctx context.Context, in *keys.SignRequest) (*keys.SignResponse, error) {
	var addr crypto.Address
	var err error
	if in.Address != nil {
		addr = *in.Address
	} else {
		addr, err = p.keys.GetName(in.GetName())
		if err != nil {
			return nil, err
		}
	}

	key, err := p.keys.GetKey(in.GetPassphrase(), addr)
	if err != nil {
		return nil, err
	}

	sig, err := key.PrivateKey.Sign(in.GetMessage())
	if err != nil {
		return nil, err
	}
	return &keys.SignResponse{Signature: *sig}, err
}

func (p *Proxy) Verify(ctx context.Context, in *keys.VerifyRequest) (*keys.VerifyResponse, error) {
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

func (p *Proxy) Hash(ctx context.Context, in *keys.HashRequest) (*keys.HashResponse, error) {
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

func (p *Proxy) ImportJSON(ctx context.Context, in *keys.ImportJSONRequest) (*keys.ImportResponse, error) {
	addr, err := p.keys.ImportJSON(in.GetPassphrase(), in.GetJSON())
	return &keys.ImportResponse{Address: *addr}, err
}

func (p *Proxy) Import(ctx context.Context, in *keys.ImportRequest) (*keys.ImportResponse, error) {
	curveT, err := crypto.CurveTypeFromString(in.GetCurveType())
	if err != nil {
		return nil, err
	}
	key, err := keys.NewKeyFromPriv(curveT, in.GetKeyBytes())
	if err != nil {
		return nil, err
	}

	// store the new key
	if err = p.keys.StoreKey(in.GetPassphrase(), key); err != nil {
		return nil, err
	}

	if in.GetName() != "" {
		err = p.keys.AddName(in.GetName(), key.Address)
		if err != nil {
			return nil, err
		}
	}
	return &keys.ImportResponse{Address: key.Address}, nil
}

func (p *Proxy) List(ctx context.Context, in *keys.ListRequest) (*keys.ListResponse, error) {
	list, err := p.keys.List(in.GetKeyName())
	return &keys.ListResponse{Key: list}, err
}

func (p *Proxy) RemoveName(ctx context.Context, in *keys.RemoveNameRequest) (*keys.RemoveNameResponse, error) {
	if in.GetKeyName() == "" {
		return nil, fmt.Errorf("please specify a name")
	}

	return &keys.RemoveNameResponse{}, p.keys.RmName(in.GetKeyName())
}

func (p *Proxy) AddName(ctx context.Context, in *keys.AddNameRequest) (*keys.AddNameResponse, error) {
	if in.GetKeyname() == "" {
		return nil, fmt.Errorf("please specify a name")
	}

	return &keys.AddNameResponse{}, p.keys.AddName(in.GetKeyname(), in.Address)
}
