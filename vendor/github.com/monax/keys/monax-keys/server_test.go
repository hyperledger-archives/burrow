package keys

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/monax/keys/crypto"
)

// start the server
func init() {
	failedCh := make(chan error)
	go func() {
		err := StartServer(DefaultHost, TestPort)
		failedCh <- err
	}()
	tick := time.NewTicker(time.Second)
	select {
	case err := <-failedCh:
		fmt.Println(err)
		os.Exit(1)
	case <-tick.C:
	}
}

// tests are identical to core_test.go but through the http calls instead of the core functions
func formatForBody(args map[string]string) *bytes.Buffer {
	bd1, _ := json.Marshal(args)
	bd2 := bytes.NewBuffer([]byte(bd1))
	return bd2
}

func testServerKeygenAndPub(t *testing.T, typ string) {
	body1 := formatForBody(map[string]string{"type": typ})
	req, _ := http.NewRequest("POST", TestAddr+"/gen", body1)

	addr, errS, err := requestResponse(req)
	checkErrs(t, errS, err)

	body2 := formatForBody(map[string]string{"addr": addr})
	req, _ = http.NewRequest("POST", TestAddr+"/pub", body2)

	pub, errS, err := requestResponse(req)
	checkErrs(t, errS, err)

	pubB, _ := hex.DecodeString(pub)
	addrB, _ := hex.DecodeString(addr)
	if err := checkAddrFromPub(typ, pubB, addrB); err != nil {
		t.Fatal(err)
	}
}

func TestServerKeygenAndPub(t *testing.T) {
	for _, typ := range KEY_TYPES {
		testServerKeygenAndPub(t, typ)
	}
}

func testServerSignAndVerify(t *testing.T, typ string) {
	body1 := formatForBody(map[string]string{"type": typ})
	req, _ := http.NewRequest("POST", TestAddr+"/gen", body1)
	addr, errS, err := requestResponse(req)
	checkErrs(t, errS, err)

	body1 = formatForBody(map[string]string{"type": typ, "addr": addr})
	req, _ = http.NewRequest("POST", TestAddr+"/pub", body1)
	pub, errS, err := requestResponse(req)
	checkErrs(t, errS, err)

	hash := crypto.Sha3([]byte("the hash of something!"))

	body2 := formatForBody(map[string]string{"msg": toHex(hash), "addr": addr})
	req, _ = http.NewRequest("POST", TestAddr+"/sign", body2)
	sig, errS, err := requestResponse(req)
	checkErrs(t, errS, err)

	body3 := formatForBody(map[string]string{"type": typ, "msg": toHex(hash), "pub": pub, "sig": sig})
	req, _ = http.NewRequest("POST", TestAddr+"/verify", body3)
	res, errS, err := requestResponse(req)
	checkErrs(t, errS, err)

	if res != "true" {
		t.Fatalf("Signature (type %s) failed to verify.\nResponse: %s\nSig %s, Hash %s, Addr %s", typ, res, sig, toHex(hash), addr)
	}
}

func TestServerSignAndVerify(t *testing.T) {
	for _, typ := range KEY_TYPES {
		testServerSignAndVerify(t, typ)
	}
}

func testServerHash(t *testing.T, typ string) {
	hData := hashData[typ]
	data, expected := hData.data, hData.expected

	body1 := formatForBody(map[string]string{"type": typ, "msg": data})
	req, _ := http.NewRequest("POST", TestAddr+"/hash", body1)
	hash, errS, err := requestResponse(req)
	checkErrs(t, errS, err)

	if hash != expected {
		t.Fatalf("Hash error for %s. Got %s, expected %s", typ, hash, expected)
	}
}

func TestServerHash(t *testing.T) {
	for _, typ := range HASH_TYPES {
		testServerHash(t, typ)
	}
}

//---------------------------------------------------------------------------------

func checkErrs(t *testing.T, errS string, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if errS != "" {
		t.Fatal(errS)
	}
}
