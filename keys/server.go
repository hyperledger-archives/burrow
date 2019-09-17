package keys

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"

	"github.com/hyperledger/burrow/crypto"
	hex "github.com/tmthrgd/go-hex"
	"golang.org/x/crypto/ripemd160"
	"google.golang.org/grpc"
)

func StandAloneServer(keysDir string, AllowBadFilePermissions bool) *grpc.Server {
	grpcServer := grpc.NewServer()
	RegisterKeysServer(grpcServer, NewKeyStore(keysDir, AllowBadFilePermissions))
	return grpcServer
}

//------------------------------------------------------------------------
// handlers

func (p *KeyStore) GenerateKey(ctx context.Context, in *GenRequest) (*GenResponse, error) {
	curveT, err := crypto.CurveTypeFromString(in.CurveType)
	if err != nil {
		return nil, err
	}

	key, err := p.Gen(in.Passphrase, curveT)
	if err != nil {
		return nil, fmt.Errorf("error generating key %s %s", curveT, err)
	}

	if in.KeyName != "" {
		err = p.SetName(in.KeyName, key.Address)
		if err != nil {
			return nil, err
		}
	}

	return &GenResponse{Address: key.Address}, nil
}

func (p *KeyStore) PublicKey(ctx context.Context, in *PubRequest) (*PubResponse, error) {
	var addr crypto.Address
	var err error

	if in.Address != nil {
		addr = *in.Address
	} else {
		addr, err = p.GetName(in.GetName())
		if err != nil {
			return nil, err
		}
	}

	// No phrase needed for public key. I hope.
	key, err := p.GetKey("", addr)
	if key == nil {
		return nil, err
	}

	return &PubResponse{CurveType: key.CurveType.String(), PublicKey: key.Pubkey()}, nil
}

func (p *KeyStore) Sign(ctx context.Context, in *SignRequest) (*SignResponse, error) {
	var addr crypto.Address
	var err error

	if in.Address != nil {
		addr = *in.Address
	} else {
		addr, err = p.GetName(in.GetName())
		if err != nil {
			return nil, err
		}
	}

	key, err := p.GetKey(in.GetPassphrase(), addr)
	if err != nil {
		return nil, err
	}

	sig, err := key.PrivateKey.Sign(in.GetMessage())
	if err != nil {
		return nil, err
	}
	return &SignResponse{Signature: *sig}, err
}

func (p *KeyStore) Verify(ctx context.Context, in *VerifyRequest) (*VerifyResponse, error) {
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

	return &VerifyResponse{}, nil
}

func (p *KeyStore) Hash(ctx context.Context, in *HashRequest) (*HashResponse, error) {
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

	return &HashResponse{Hash: hex.EncodeUpperToString(hasher.Sum(nil))}, nil
}

func (p *KeyStore) ImportJSON(ctx context.Context, in *ImportJSONRequest) (*ImportResponse, error) {
	keyJSON := []byte(in.GetJSON())
	addr, err := IsValidKeyJson(keyJSON)
	if err == nil {
		err = p.StoreKeyRaw(addr, keyJSON)
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

		addr, err = crypto.AddressFromHexString(j1.Address)
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

		key, err := NewKeyFromPriv(curveT, privKey)
		if err != nil {
			return nil, err
		}

		// store the new key
		if err = p.StoreKey(in.GetPassphrase(), key); err != nil {
			return nil, err
		}
	}
	return &ImportResponse{Address: addr}, nil
}

func (p *KeyStore) Import(ctx context.Context, in *ImportRequest) (*ImportResponse, error) {
	curveT, err := crypto.CurveTypeFromString(in.GetCurveType())
	if err != nil {
		return nil, err
	}
	key, err := NewKeyFromPriv(curveT, in.GetKeyBytes())
	if err != nil {
		return nil, err
	}

	// store the new key
	if err = p.StoreKey(in.GetPassphrase(), key); err != nil {
		return nil, err
	}

	if in.GetName() != "" {
		err = p.SetName(in.GetName(), key.Address)
		if err != nil {
			return nil, err
		}
	}
	return &ImportResponse{Address: key.Address}, nil
}

func (p *KeyStore) List(ctx context.Context, in *ListRequest) (*ListResponse, error) {
	byname, err := p.GetAllNames()
	if err != nil {
		return nil, err
	}

	var list []*KeyID

	if in.KeyName != "" {
		if addr, ok := byname[in.KeyName]; ok {
			list = append(list, &KeyID{KeyName: getAddressNames(addr, byname), Address: addr})
		} else {
			if addr, err := crypto.AddressFromHexString(in.KeyName); err == nil {
				_, err := p.GetKey("", addr)
				if err == nil {
					list = append(list, &KeyID{Address: addr, KeyName: getAddressNames(addr, byname)})
				}
			}
		}
	} else {
		// list all address
		addrs, err := p.GetAllAddresses()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			list = append(list, &KeyID{KeyName: getAddressNames(addr, byname), Address: addr})
		}
	}

	return &ListResponse{Key: list}, nil
}

func getAddressNames(address crypto.Address, byname map[string]crypto.Address) []string {
	names := make([]string, 0)

	for name, addr := range byname {
		if address == addr {
			names = append(names, name)
		}
	}

	return names
}

func (p *KeyStore) RemoveName(ctx context.Context, in *RemoveNameRequest) (*RemoveNameResponse, error) {
	if in.GetKeyName() == "" {
		return nil, fmt.Errorf("please specify a name")
	}

	return &RemoveNameResponse{}, p.RmName(in.GetKeyName())
}

func (p *KeyStore) AddName(ctx context.Context, in *AddNameRequest) (*AddNameResponse, error) {
	if in.GetKeyname() == "" {
		return nil, fmt.Errorf("please specify a name")
	}

	return &AddNameResponse{}, p.SetName(in.GetKeyname(), in.Address)
}
