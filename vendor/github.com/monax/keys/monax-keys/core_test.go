package keys

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/monax/keys/common"
	"github.com/monax/keys/crypto"
	ed25519 "github.com/monax/keys/crypto/helpers"
)

var (
	AUTH       = ""
	KEY_TYPES  = []string{"secp256k1,sha3", "ed25519,ripemd160", "secp256k1,ripemd160sha256"}
	HASH_TYPES = []string{"sha256", "ripemd160"}
)

func init() {
	// TODO: randomize and do setup/tear down for tests
	KeysDir = common.ScratchPath
}

func dumpKey(k *crypto.Key) {
	b, _ := k.MarshalJSON()
	fmt.Println(string(b))
}

func testKeygenAndPub(t *testing.T, typ string) {
	addr, err := coreKeygen(AUTH, typ)
	if err != nil {
		t.Fatal(err)
	}

	pub, err := corePub(toHex(addr))
	if err != nil {
		t.Fatal(err)
	}

	if err := checkAddrFromPub(typ, pub, addr); err != nil {
		t.Fatal(err)
	}

}

func TestKeygenAndPub(t *testing.T) {
	for _, typ := range KEY_TYPES {
		testKeygenAndPub(t, typ)
	}
}

func testSignAndVerify(t *testing.T, typ string) {
	addr, err := coreKeygen(AUTH, typ)
	if err != nil {
		t.Fatal(err)
	}

	pub, err := corePub(hex.EncodeToString(addr))
	if err != nil {
		t.Fatal(err)
	}

	hash := crypto.Sha3([]byte("the hash of something!"))

	sig, err := coreSign(toHex(hash), toHex(addr))
	if err != nil {
		t.Fatal(err)
	}

	res, err := coreVerify(typ, toHex(pub), toHex(hash), toHex(sig))
	if err != nil {
		t.Fatal(err)
	}
	if res != true {
		t.Fatalf("Signature (type %s) failed to verify.\nResponse: %v\nSig %x, Hash %x, Addr %x", typ, res, sig, hash, addr)
	}
}

func TestSignAndVerify(t *testing.T) {
	for _, typ := range KEY_TYPES {
		testSignAndVerify(t, typ)
	}
}

func testHash(t *testing.T, typ string) {
	hData := hashData[typ]
	data, expected := hData.data, hData.expected
	hash, err := coreHash(typ, data, false)
	if err != nil {
		t.Fatal(err)
	}

	if toHex(hash) != expected {
		t.Fatalf("Hash error for %s. Got %s, expected %s", typ, toHex(hash), expected)
	}

}

type hashInfo struct {
	data     string
	expected string
}

var hashData = map[string]hashInfo{
	"sha256":    {"hi", "8F434346648F6B96DF89DDA901C5176B10A6D83961DD3C1AC88B59B2DC327AA4"},
	"ripemd160": {"hi", "242485AB6BFD3502BCB3442EA2E211687B8E4D89"},
}

func TestHash(t *testing.T) {
	for _, typ := range HASH_TYPES {
		testHash(t, typ)
	}
}

//--------------------------------------------------------------------------------

func toHex(b []byte) string {
	return fmt.Sprintf("%X", b)
}

func checkAddrFromPub(typ string, pub, addr []byte) error {
	var addr2 []byte
	switch typ {
	case "secp256k1,sha3":
		addr2 = crypto.Sha3(pub[1:])[12:]
	case "secp256k1,ripemd160sha256":
		addr2 = crypto.Ripemd160(crypto.Sha256(pub))
	case "ed25519,ripemd160":
		// XXX: something weird here. I have seen this oscillate!
		// addr2 = binary.BinaryRipemd160(pub)
		var pubArray ed25519.PubKeyEd25519
		copy(pubArray[:], pub)
		addr2 = pubArray.Address()
	default:
		return fmt.Errorf("Unknown or incomplete typ %s", typ)
	}
	if bytes.Compare(addr, addr2) != 0 {
		return fmt.Errorf("Keygen addr doesn't match pub. Got %X, expected %X", addr2, addr)
	}
	return nil
}
