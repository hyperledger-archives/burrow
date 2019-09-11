package proxy

import (
	"context"
	"crypto/sha256"
	"encoding/json"
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

	return &keys.GenResponse{Address: key.Address.String()}, nil
}

func (p *Proxy) Export(ctx context.Context, in *keys.ExportRequest) (*keys.ExportResponse, error) {
	addr, err := p.keys.GetNameAddr(in.GetName(), in.GetAddress())
	if err != nil {
		return nil, err
	}

	// No phrase needed for public key. I hope.
	key, err := p.keys.GetKey(in.GetPassphrase(), addr.Bytes())
	if err != nil {
		return nil, err
	}

	return &keys.ExportResponse{
		Address:    addr.Bytes(),
		CurveType:  key.CurveType.String(),
		Publickey:  key.PublicKey.PublicKey[:],
		Privatekey: key.PrivateKey.PrivateKey[:],
	}, nil
}

func (p *Proxy) PublicKey(ctx context.Context, in *keys.PubRequest) (*keys.PubResponse, error) {
	addr, err := p.keys.GetNameAddr(in.GetName(), in.GetAddress())
	if err != nil {
		return nil, err
	}

	// No phrase needed for public key. I hope.
	key, err := p.keys.GetKey("", addr.Bytes())
	if key == nil {
		return nil, err
	}

	return &keys.PubResponse{CurveType: key.CurveType.String(), PublicKey: key.Pubkey()}, nil
}

func (p *Proxy) Sign(ctx context.Context, in *keys.SignRequest) (*keys.SignResponse, error) {
	addr, err := p.keys.GetNameAddr(in.GetName(), in.GetAddress())
	if err != nil {
		return nil, err
	}

	key, err := p.keys.GetKey(in.GetPassphrase(), addr.Bytes())
	if err != nil {
		return nil, err
	}

	sig, err := key.PrivateKey.Sign(in.GetMessage())
	if err != nil {
		return nil, err
	}
	return &keys.SignResponse{Signature: sig}, err
}

func (p *Proxy) Verify(ctx context.Context, in *keys.VerifyRequest) (*keys.VerifyResponse, error) {
	if in.GetPublicKey() == nil {
		return nil, fmt.Errorf("must provide a pubkey")
	}
	if in.GetMessage() == nil {
		return nil, fmt.Errorf("must provide a message")
	}
	if in.GetSignature() == nil {
		return nil, fmt.Errorf("must provide a signature")
	}

	sig := in.GetSignature()
	pubkey, err := crypto.PublicKeyFromBytes(in.GetPublicKey(), sig.GetCurveType())
	if err != nil {
		return nil, err
	}
	err = pubkey.Verify(in.GetMessage(), sig)
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
	keyJSON := []byte(in.GetJSON())
	addr := keys.IsValidKeyJson(keyJSON)
	if addr != nil {
		addr, err := crypto.AddressFromBytes(addr)
		if err != nil {
			return nil, err
		}
		err = p.keys.StoreKeyRaw(addr, keyJSON)
		if err != nil {
			return nil, err
		}
	} else {
		j1 := new(struct {
			CurveType   string
			Address     string
			PublicKey   string
			AddressHash string
			PrivateKey  string
		})

		err := json.Unmarshal([]byte(in.GetJSON()), &j1)
		if err != nil {
			return nil, err
		}

		addr, err = hex.DecodeString(j1.Address)
		if err != nil {
			return nil, err
		}

		curveT, err := crypto.CurveTypeFromString(j1.CurveType)
		if err != nil {
			return nil, err
		}

		privKey, err := hex.DecodeString(j1.PrivateKey)
		if err != nil {
			return nil, err
		}

		key, err := keys.NewKeyFromPriv(curveT, privKey)
		if err != nil {
			return nil, err
		}

		// store the new key
		if err = p.keys.StoreKey(in.GetPassphrase(), key); err != nil {
			return nil, err
		}
	}
	return &keys.ImportResponse{Address: hex.EncodeUpperToString(addr)}, nil
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
	return &keys.ImportResponse{Address: key.Address.String()}, nil
}

func (p *Proxy) List(ctx context.Context, in *keys.ListRequest) (*keys.ListResponse, error) {
	byname, err := p.keys.GetAllNames()
	if err != nil {
		return nil, err
	}

	var list []*keys.KeyID

	if in.KeyName != "" {
		if addr, ok := byname[in.KeyName]; ok {
			list = append(list, &keys.KeyID{KeyName: getAddressNames(addr, byname), Address: addr})
		} else {
			if addr, err := crypto.AddressFromHexString(in.KeyName); err == nil {
				_, err := p.keys.GetKey("", addr.Bytes())
				if err == nil {
					address := addr.String()
					list = append(list, &keys.KeyID{Address: address, KeyName: getAddressNames(address, byname)})
				}
			}
		}
	} else {
		// list all address
		addrs, err := p.keys.GetAllAddresses()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			list = append(list, &keys.KeyID{KeyName: getAddressNames(addr, byname), Address: addr})
		}
	}

	return &keys.ListResponse{Key: list}, nil
}

func getAddressNames(address string, byname map[string]string) []string {
	names := make([]string, 0)

	for name, addr := range byname {
		if address == addr {
			names = append(names, name)
		}
	}

	return names
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

	addr, err := crypto.AddressFromHexString(in.GetAddress())
	if err != nil {
		return nil, err
	}

	return &keys.AddNameResponse{}, p.keys.AddName(in.GetKeyname(), addr)
}
