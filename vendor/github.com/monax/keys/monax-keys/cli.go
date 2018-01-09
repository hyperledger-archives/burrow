package keys

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	. "github.com/monax/keys/common"
	//"github.com/howeyc/gopass"
	"github.com/spf13/cobra"
)

func ExitConnectErr(err error) {
	Exit(fmt.Errorf("Could not connect to monax-keys server. Start it with `monax-keys server &`. Error: %v", err))
}

func cliServer(cmd *cobra.Command, args []string) {
	IfExit(StartServer(KeyHost, KeyPort))
}

func cliKeygen(cmd *cobra.Command, args []string) {
	var auth string
	if !NoPassword {
		auth = hiddenAuth()
	}

	r, err := Call("gen", map[string]string{"auth": auth, "type": KeyType, "name": KeyName})
	if _, ok := err.(ErrConnectionRefused); ok {
		ExitConnectErr(err)
	}
	LogToChannel([]byte(r))
}

func cliLock(cmd *cobra.Command, args []string) {
	r, err := Call("lock", map[string]string{"addr": KeyAddr, "name": KeyName})
	if _, ok := err.(ErrConnectionRefused); ok {
		ExitConnectErr(err)
	}
	IfExit(err)
	LogToChannel([]byte(r))
}

func cliConvert(cmd *cobra.Command, args []string) {
	r, err := Call("mint", map[string]string{"addr": KeyAddr, "name": KeyName})
	if _, ok := err.(ErrConnectionRefused); ok {
		ExitConnectErr(err)
	}
	IfExit(err)
	LogToChannel([]byte(r))
}

func cliUnlock(cmd *cobra.Command, args []string) {
	auth := hiddenAuth()
	r, err := Call("unlock", map[string]string{"auth": auth, "addr": KeyAddr, "name": KeyName})
	if _, ok := err.(ErrConnectionRefused); ok {
		ExitConnectErr(err)
	}
	IfExit(err)
	LogToChannel([]byte(r))
}

// since pubs are not saved, the key needs to be unlocked to get the pub
// TODO: save the pubkey (backwards compatibly...)
func cliPub(cmd *cobra.Command, args []string) {
	r, err := Call("pub", map[string]string{"addr": KeyAddr, "name": KeyName})
	if _, ok := err.(ErrConnectionRefused); ok {
		ExitConnectErr(err)
	}
	IfExit(err)
	LogToChannel([]byte(r))
}

func cliSign(cmd *cobra.Command, args []string) {
	_, addr, name := KeysDir, KeyAddr, KeyName
	if len(args) != 1 {
		Exit(fmt.Errorf("enter a msg/hash to sign"))
	}
	msg := args[0]
	r, err := Call("sign", map[string]string{"addr": addr, "name": name, "msg": msg})
	if _, ok := err.(ErrConnectionRefused); ok {
		ExitConnectErr(err)
	}
	IfExit(err)
	LogToChannel([]byte(r))
}

func cliVerify(cmd *cobra.Command, args []string) {
	if len(args) != 3 {
		Exit(fmt.Errorf("enter a msg/hash, a signature, and a public key"))
	}
	msg, sig, pub := args[0], args[1], args[2]
	r, err := Call("verify", map[string]string{"type": KeyType, "pub": pub, "msg": msg, "sig": sig})
	if _, ok := err.(ErrConnectionRefused); ok {
		ExitConnectErr(err)
	}
	IfExit(err)
	LogToChannel([]byte(r))
}

func cliHash(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		Exit(fmt.Errorf("enter something to hash"))
	}
	msg := args[0]
	r, err := Call("hash", map[string]string{"type": HashType, "msg": msg, "hex": fmt.Sprintf("%v", HexByte)})
	if _, ok := err.(ErrConnectionRefused); ok {
		ExitConnectErr(err)
	}
	IfExit(err)
	LogToChannel([]byte(r))
}

func cliImport(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		Exit(fmt.Errorf("enter a private key, filename, or raw json"))
	}

	key := args[0]

	// if the key is a path, read it
	if _, err := os.Stat(key); err == nil {
		keyBytes, err := ioutil.ReadFile(key)
		key = string(keyBytes)
		IfExit(err)
	}

	var auth string
	if !NoPassword {
		log.Printf("Warning: Please note that this encryption will only take effect if you passed a raw private key (TODO!).")
		auth = hiddenAuth()
	}

	r, err := Call("import", map[string]string{"auth": auth, "name": KeyName, "type": KeyType, "key": key})
	if _, ok := err.(ErrConnectionRefused); ok {
		ExitConnectErr(err)
	}
	IfExit(err)
	LogToChannel([]byte(r))
}

func cliName(cmd *cobra.Command, args []string) {
	var name, addr string
	if len(args) > 0 {
		name = args[0]
	}
	if len(args) > 1 {
		addr = args[1]
	}

	r, err := Call("name", map[string]string{"name": name, "addr": addr})
	if _, ok := err.(ErrConnectionRefused); ok {
		ExitConnectErr(err)
	}
	IfExit(err)
	LogToChannel([]byte(r))
}

func cliNameLs(cmd *cobra.Command, args []string) {
	r, err := Call("name/ls", map[string]string{})
	if _, ok := err.(ErrConnectionRefused); ok {
		ExitConnectErr(err)
	}
	IfExit(err)
	names := make(map[string]string)
	IfExit(json.Unmarshal([]byte(r), &names))
	for n, a := range names {
		log.Printf("%s: %s\n", n, a)
	}
	LogToChannel([]byte(r))
}

func cliNameRm(cmd *cobra.Command, args []string) {
	var name string
	if len(args) > 0 {
		name = args[0]
	}
	r, err := Call("name/rm", map[string]string{"name": name})
	if _, ok := err.(ErrConnectionRefused); ok {
		ExitConnectErr(err)
	}
	IfExit(err)
	LogToChannel([]byte(r))
}
