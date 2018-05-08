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
	"golang.org/x/crypto/ripemd160"
	"google.golang.org/grpc"
)

//------------------------------------------------------------------------
// all cli commands pass through the http server
// the server process also maintains the unlocked accounts

type server struct{}

var GlobalKeyServer server

func startServer() error {
	if GlobalKeystore == nil {
		ks, err := newKeyStore()
		if err != nil {
			return err
		}

		GlobalKeystore = ks
	}
	GlobalKeyServer = server{}

	return nil
}

func StartGRPCServer(grpcserver *grpc.Server, keyConfig *KeysConfig) error {
	err := startServer()
	if err != nil {
		return err
	}
	if keyConfig.ServerEnabled {
		pbkeys.RegisterKeysServer(grpcserver, &GlobalKeyServer)
	}
	return nil
}

func StartStandAloneServer(host, port string) error {
	err := startServer()
	if err != nil {
		return err
	}

	listen, err := net.Listen("tcp", host+":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	pbkeys.RegisterKeysServer(grpcServer, &server{})
	return grpcServer.Serve(listen)
}

//------------------------------------------------------------------------
// handlers

func (k *server) GenerateKey(ctx context.Context, in *pbkeys.GenRequest) (*pbkeys.GenResponse, error) {
	curveT, err := crypto.CurveTypeFromString(in.Curvetype)
	if err != nil {
		return nil, err
	}

	key, err := GlobalKeystore.GenerateKey(in.Passphrase, curveT)
	if err != nil {
		return nil, fmt.Errorf("error generating key %s %s", curveT, err)
	}

	addrH := key.Address.String()
	if in.Keyname != "" {
		err = coreNameAdd(in.Keyname, addrH)
		if err != nil {
			return nil, err
		}
	}

	return &pbkeys.GenResponse{Address: addrH}, nil
}

func (k *server) Export(ctx context.Context, in *pbkeys.ExportRequest) (*pbkeys.ExportResponse, error) {
	addr, err := getNameAddr(in.GetName(), in.GetAddress())

	if err != nil {
		return nil, err
	}

	resp, err := coreExport(in.GetPassphrase(), addr)
	if err != nil {
		return nil, err
	}

	return &pbkeys.ExportResponse{Export: string(resp)}, nil
}

func (k *server) PublicKey(ctx context.Context, in *pbkeys.PubRequest) (*pbkeys.PubResponse, error) {
	addr, err := getNameAddr(in.GetName(), in.GetAddress())
	if err != nil {
		return nil, err
	}

	addrB, err := crypto.AddressFromHexString(addr)
	if err != nil {
		return nil, err
	}

	// No phrase needed for public key. I hope.
	key, err := GlobalKeystore.GetKey("", addrB.Bytes())
	if key == nil {
		return nil, err
	}

	return &pbkeys.PubResponse{Curvetype: key.CurveType.String(), Pub: key.Pubkey()}, nil
}

func (k *server) Sign(ctx context.Context, in *pbkeys.SignRequest) (*pbkeys.SignResponse, error) {
	addr, err := crypto.AddressFromHexString(in.Address)
	if err != nil {
		return nil, err
	}

	key, err := GlobalKeystore.GetKey(in.GetPassphrase(), addr[:])
	if err != nil {
		return nil, err
	}

	sig, err := key.Sign(in.GetHash())

	return &pbkeys.SignResponse{Signature: sig}, nil
}

func (k *server) Verify(ctx context.Context, in *pbkeys.VerifyRequest) (*pbkeys.Empty, error) {
	if in.GetPub() == nil {
		return nil, fmt.Errorf("must provide a pubkey")
	}
	if in.GetHash() == nil {
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
	match := pubkey.Verify(in.GetHash(), sig)
	if !match {
		return nil, fmt.Errorf("Signature does not match")
	}

	return &pbkeys.Empty{}, nil
}

func (k *server) Hash(ctx context.Context, in *pbkeys.HashRequest) (*pbkeys.HashResponse, error) {
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

	return &pbkeys.HashResponse{Hash: fmt.Sprintf("%X", hasher.Sum(nil))}, nil
}

func (k *server) ImportJSON(ctx context.Context, in *pbkeys.ImportJSONRequest) (*pbkeys.ImportResponse, error) {
	keyJSON := []byte(in.GetJSON())
	var err error
	addr := IsValidKeyJson(keyJSON)
	if addr != nil {
		_, err = writeKey(KeysDir, addr, keyJSON)
	} else {
		err = fmt.Errorf("invalid json key passed on command line")
	}
	if err != nil {
		return nil, err
	}
	return &pbkeys.ImportResponse{Address: fmt.Sprintf("%X", addr)}, nil
}

func (k *server) Import(ctx context.Context, in *pbkeys.ImportRequest) (*pbkeys.ImportResponse, error) {
	curveT, err := crypto.CurveTypeFromString(in.GetCurvetype())
	if err != nil {
		return nil, err
	}
	key, err := NewKeyFromPriv(curveT, in.GetKeybytes())
	if err != nil {
		return nil, err
	}

	// store the new key
	if err = GlobalKeystore.StoreKey(in.GetPassphrase(), key); err != nil {
		return nil, err
	}

	if in.GetName() != "" {
		if err := coreNameAdd(in.GetName(), key.Address.String()); err != nil {
			return nil, err
		}
	}
	return &pbkeys.ImportResponse{Address: fmt.Sprintf("%X", key.Address)}, nil
}

func (k *server) List(ctx context.Context, in *pbkeys.Name) (*pbkeys.ListResponse, error) {
	names, err := coreNameList()
	if err != nil {
		return nil, err
	}

	var list []*pbkeys.Key

	for name, addr := range names {
		list = append(list, &pbkeys.Key{Keyname: name, Address: addr})
	}

	return &pbkeys.ListResponse{Key: list}, nil
}

func (k *server) Remove(ctx context.Context, in *pbkeys.Name) (*pbkeys.Empty, error) {
	if in.GetKeyname() == "" {
		return nil, fmt.Errorf("please specify a name")
	}

	return &pbkeys.Empty{}, coreNameRm(in.GetKeyname())
}

func (k *server) Add(ctx context.Context, in *pbkeys.AddRequest) (*pbkeys.Empty, error) {
	if in.GetKeyname() == "" {
		return nil, fmt.Errorf("please specify a name")
	}

	if in.GetAddress() == "" {
		return nil, fmt.Errorf("please specify an address")
	}

	return &pbkeys.Empty{}, coreNameAdd(in.GetKeyname(), strings.ToUpper(in.GetAddress()))
}
