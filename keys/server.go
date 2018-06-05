package keys

import (
	"context"
	"crypto/sha256"
	"fmt"
	"hash"
	"net"
	"strings"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/keys/pbkeys"
	"github.com/hyperledger/burrow/logging"
	"github.com/tmthrgd/go-hex"
	"golang.org/x/crypto/ripemd160"
	"google.golang.org/grpc"
)

//------------------------------------------------------------------------
// all cli commands pass through the http KeyStore
// the KeyStore process also maintains the unlocked accounts

func StartStandAloneServer(keysDir, host, port string, AllowBadFilePermissions bool, logger *logging.Logger) error {
	listen, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		return err
	}

	ks := NewKeyStore(keysDir, AllowBadFilePermissions, logger)

	grpcServer := grpc.NewServer()
	pbkeys.RegisterKeysServer(grpcServer, &ks)
	return grpcServer.Serve(listen)
}

//------------------------------------------------------------------------
// handlers

func (k *KeyStore) GenerateKey(ctx context.Context, in *pbkeys.GenRequest) (*pbkeys.GenResponse, error) {
	curveT, err := crypto.CurveTypeFromString(in.Curvetype)
	if err != nil {
		return nil, err
	}

	key, err := k.Gen(in.Passphrase, curveT)
	if err != nil {
		return nil, fmt.Errorf("error generating key %s %s", curveT, err)
	}

	addrH := key.Address.String()
	if in.Keyname != "" {
		err = coreNameAdd(k.keysDirPath, in.Keyname, addrH)
		if err != nil {
			return nil, err
		}
	}

	return &pbkeys.GenResponse{Address: addrH}, nil
}

func (k *KeyStore) Export(ctx context.Context, in *pbkeys.ExportRequest) (*pbkeys.ExportResponse, error) {
	addr, err := getNameAddr(k.keysDirPath, in.GetName(), in.GetAddress())

	if err != nil {
		return nil, err
	}

	addrB, err := crypto.AddressFromHexString(addr)
	if err != nil {
		return nil, err
	}

	// No phrase needed for public key. I hope.
	key, err := k.GetKey(in.GetPassphrase(), addrB.Bytes())
	if err != nil {
		return nil, err
	}
	resp, err := coreExport(key)
	if err != nil {
		return nil, err
	}

	return &pbkeys.ExportResponse{Export: string(resp)}, nil
}

func (k *KeyStore) PublicKey(ctx context.Context, in *pbkeys.PubRequest) (*pbkeys.PubResponse, error) {
	addr, err := getNameAddr(k.keysDirPath, in.GetName(), in.GetAddress())
	if err != nil {
		return nil, err
	}

	addrB, err := crypto.AddressFromHexString(addr)
	if err != nil {
		return nil, err
	}

	// No phrase needed for public key. I hope.
	key, err := k.GetKey("", addrB.Bytes())
	if key == nil {
		return nil, err
	}

	return &pbkeys.PubResponse{Curvetype: key.CurveType.String(), Pub: key.Pubkey()}, nil
}

func (k *KeyStore) Sign(ctx context.Context, in *pbkeys.SignRequest) (*pbkeys.SignResponse, error) {
	addr, err := getNameAddr(k.keysDirPath, in.GetName(), in.GetAddress())
	if err != nil {
		return nil, err
	}

	addrB, err := crypto.AddressFromHexString(addr)
	if err != nil {
		return nil, err
	}

	key, err := k.GetKey(in.GetPassphrase(), addrB[:])
	if err != nil {
		return nil, err
	}

	sig, err := key.Sign(in.GetMessage())

	return &pbkeys.SignResponse{Signature: sig, Curvetype: key.CurveType.String()}, nil
}

func (k *KeyStore) Verify(ctx context.Context, in *pbkeys.VerifyRequest) (*pbkeys.Empty, error) {
	if in.GetPub() == nil {
		return nil, fmt.Errorf("must provide a pubkey")
	}
	if in.GetMessage() == nil {
		return nil, fmt.Errorf("must provide a message")
	}
	if in.GetSignature() == nil {
		return nil, fmt.Errorf("must provide a signature")
	}

	curveT, err := crypto.CurveTypeFromString(in.GetCurvetype())
	if err != nil {
		return nil, err
	}
	sig, err := crypto.SignatureFromBytes(in.GetSignature(), curveT)
	if err != nil {
		return nil, err
	}
	pubkey, err := crypto.PublicKeyFromBytes(in.GetPub(), curveT)
	if err != nil {
		return nil, err
	}
	match := pubkey.Verify(in.GetMessage(), sig)
	if !match {
		return nil, fmt.Errorf("Signature does not match")
	}

	return &pbkeys.Empty{}, nil
}

func (k *KeyStore) Hash(ctx context.Context, in *pbkeys.HashRequest) (*pbkeys.HashResponse, error) {
	var hasher hash.Hash
	switch in.GetHashtype() {
	case "ripemd160":
		hasher = ripemd160.New()
	case "sha256":
		hasher = sha256.New()
	// case "sha3":
	default:
		return nil, fmt.Errorf("Unknown hash type %v", in.GetHashtype())
	}

	hasher.Write(in.GetMessage())

	return &pbkeys.HashResponse{Hash: hex.EncodeUpperToString(hasher.Sum(nil))}, nil
}

func (k *KeyStore) ImportJSON(ctx context.Context, in *pbkeys.ImportJSONRequest) (*pbkeys.ImportResponse, error) {
	keyJSON := []byte(in.GetJSON())
	var err error
	addr := IsValidKeyJson(keyJSON)
	if addr != nil {
		_, err = writeKey(k.keysDirPath, addr, keyJSON)
	} else {
		err = fmt.Errorf("invalid json key passed on command line")
	}
	if err != nil {
		return nil, err
	}
	return &pbkeys.ImportResponse{Address: hex.EncodeUpperToString(addr)}, nil
}

func (k *KeyStore) Import(ctx context.Context, in *pbkeys.ImportRequest) (*pbkeys.ImportResponse, error) {
	curveT, err := crypto.CurveTypeFromString(in.GetCurvetype())
	if err != nil {
		return nil, err
	}
	key, err := NewKeyFromPriv(curveT, in.GetKeybytes())
	if err != nil {
		return nil, err
	}

	// store the new key
	if err = k.StoreKey(in.GetPassphrase(), key); err != nil {
		return nil, err
	}

	if in.GetName() != "" {
		if err := coreNameAdd(k.keysDirPath, in.GetName(), key.Address.String()); err != nil {
			return nil, err
		}
	}
	return &pbkeys.ImportResponse{Address: hex.EncodeUpperToString(key.Address[:])}, nil
}

func (k *KeyStore) List(ctx context.Context, in *pbkeys.Name) (*pbkeys.ListResponse, error) {
	names, err := coreNameList(k.keysDirPath)
	if err != nil {
		return nil, err
	}

	var list []*pbkeys.Key

	for name, addr := range names {
		list = append(list, &pbkeys.Key{Keyname: name, Address: addr})
	}

	return &pbkeys.ListResponse{Key: list}, nil
}

func (k *KeyStore) RemoveName(ctx context.Context, in *pbkeys.Name) (*pbkeys.Empty, error) {
	if in.GetKeyname() == "" {
		return nil, fmt.Errorf("please specify a name")
	}

	return &pbkeys.Empty{}, coreNameRm(k.keysDirPath, in.GetKeyname())
}

func (k *KeyStore) AddName(ctx context.Context, in *pbkeys.AddNameRequest) (*pbkeys.Empty, error) {
	if in.GetKeyname() == "" {
		return nil, fmt.Errorf("please specify a name")
	}

	if in.GetAddress() == "" {
		return nil, fmt.Errorf("please specify an address")
	}

	return &pbkeys.Empty{}, coreNameAdd(k.keysDirPath, in.GetKeyname(), strings.ToUpper(in.GetAddress()))
}
