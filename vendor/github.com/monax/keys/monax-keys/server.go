package keys

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/rs/cors"
)

//------------------------------------------------------------------------
// all cli commands pass through the http server
// the server process also maintains the unlocked accounts

func StartServer(host, port string) error {
	ks, err := newKeyStoreAuth()
	if err != nil {
		return err
	}

	AccountManager = NewManager(ks)

	mux := http.NewServeMux()
	mux.HandleFunc("/gen", genHandler)
	mux.HandleFunc("/pub", pubHandler)
	mux.HandleFunc("/sign", signHandler)
	mux.HandleFunc("/verify", verifyHandler)
	mux.HandleFunc("/hash", hashHandler)
	mux.HandleFunc("/import", importHandler)
	mux.HandleFunc("/name", nameHandler)
	mux.HandleFunc("/name/ls", nameLsHandler)
	mux.HandleFunc("/name/rm", nameRmHandler)
	mux.HandleFunc("/unlock", unlockHandler)
	mux.HandleFunc("/lock", lockHandler)
	mux.HandleFunc("/mint", convertMintHandler)

	log.Printf("Starting monax-keys server on %s:%s\n", host, port)
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // TODO: dev
	})
	return http.ListenAndServe(host+":"+port, c.Handler(mux))
}

// A request is just a map of args to be json marshalled
type HTTPRequest map[string]string

// dead simple response struct
type HTTPResponse struct {
	Response string
	Error    string
}

func WriteResult(w http.ResponseWriter, result string) {
	resp := HTTPResponse{result, ""}
	b, _ := json.Marshal(resp)
	w.Write(b)
}

func WriteError(w http.ResponseWriter, err error) {
	resp := HTTPResponse{"", err.Error()}
	b, _ := json.Marshal(resp)
	w.Write(b)
}

//------------------------------------------------------------------------
// handlers

func genHandler(w http.ResponseWriter, r *http.Request) {
	typ, auth, args, err := typeAuthArgs(r)
	if err != nil {
		WriteError(w, err)
		return
	}

	name := args["name"]
	addr, err := coreKeygen(auth, typ)
	if err != nil {
		WriteError(w, err)
		return
	}
	if name != "" {
		err := coreNameAdd(name, strings.ToUpper(hex.EncodeToString(addr)))
		if err != nil {
			WriteError(w, err)
			return
		}
	}
	WriteResult(w, fmt.Sprintf("%X", addr))
}

func unlockHandler(w http.ResponseWriter, r *http.Request) {
	_, auth, args, err := typeAuthArgs(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	addr, name, timeout := args["addr"], args["name"], args["timeout"]
	addr, err = getNameAddr(name, addr)
	if err != nil {
		WriteError(w, err)
		return
	}
	if err := coreUnlock(auth, addr, timeout); err != nil {
		WriteError(w, err)
		return
	}
	WriteResult(w, fmt.Sprintf("%s unlocked", addr))
}

func convertMintHandler(w http.ResponseWriter, r *http.Request) {
	_, _, args, err := typeAuthArgs(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	addr, name := args["addr"], args["name"]
	addr, err = getNameAddr(name, addr)
	if err != nil {
		WriteError(w, err)
		return
	}
	key, err := coreConvert(addr)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteResult(w, string(key))
}

func lockHandler(w http.ResponseWriter, r *http.Request) {
	// TODO
}

func pubHandler(w http.ResponseWriter, r *http.Request) {
	_, _, args, err := typeAuthArgs(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	addr, name := args["addr"], args["name"]
	addr, err = getNameAddr(name, addr)
	if err != nil {
		WriteError(w, err)
		return
	}
	pub, err := corePub(addr)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteResult(w, fmt.Sprintf("%X", pub))
}

func signHandler(w http.ResponseWriter, r *http.Request) {
	_, _, args, err := typeAuthArgs(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	addr, name := args["addr"], args["name"]
	addr, err = getNameAddr(name, addr)
	if err != nil {
		WriteError(w, err)
		return
	}
	msg := args["msg"]
	if msg == "" {
		WriteError(w, fmt.Errorf("must provide a message to sign with the `msg` key"))
		return
	}
	sig, err := coreSign(msg, addr)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteResult(w, fmt.Sprintf("%X", sig))
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	typ, _, args, err := typeAuthArgs(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	pub, msg, sig := args["pub"], args["msg"], args["sig"]
	if pub == "" {
		WriteError(w, fmt.Errorf("must provide a pubkey with the `pub` key"))
		return
	}
	if msg == "" {
		WriteError(w, fmt.Errorf("must provide a message msg with the `msg` key"))
		return
	}
	if sig == "" {
		WriteError(w, fmt.Errorf("must provide a signature with the `sig` key"))
		return
	}

	res, err := coreVerify(typ, pub, msg, sig)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteResult(w, fmt.Sprintf("%v", res))
}

func hashHandler(w http.ResponseWriter, r *http.Request) {
	typ, _, args, err := typeAuthArgs(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	msg := args["msg"]
	hexD := args["hex"]

	hash, err := coreHash(typ, msg, hexD == "true")
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteResult(w, fmt.Sprintf("%X", hash))
}

func importHandler(w http.ResponseWriter, r *http.Request) {
	typ, auth, args, err := typeAuthArgs(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	name, key := args["data"], args["key"]

	addr, err := coreImport(auth, typ, key)
	if err != nil {
		WriteError(w, err)
		return
	}

	if name != "" {
		if err := coreNameAdd(name, strings.ToUpper(hex.EncodeToString(addr))); err != nil {
			WriteError(w, err)
			return
		}
	}
	WriteResult(w, fmt.Sprintf("%X", addr))
}

func nameHandler(w http.ResponseWriter, r *http.Request) {
	_, _, args, err := typeAuthArgs(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	name, addr := args["name"], args["addr"]

	// log.Debugf("name handler. name (%s). addr (%s)\n", name, addr)

	if name == "" {
		WriteError(w, fmt.Errorf("please specify a name"))
		return
	}

	if addr == "" {
		addr, err := coreNameGet(name)
		if err != nil {
			WriteError(w, err)
			return
		}
		WriteResult(w, addr)
	} else {
		if err := coreNameAdd(name, strings.ToUpper(addr)); err != nil {
			WriteError(w, err)
			return
		}
		WriteResult(w, fmt.Sprintf("Added name (%s) to address (%s)", name, addr))
	}
}

func nameLsHandler(w http.ResponseWriter, r *http.Request) {
	_, _, _, err := typeAuthArgs(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	// name, addr := args["name"], args["addr"]
	// log.Debugf("name ls handler. name (%s). addr (%s)\n", name, addr)

	names, err := coreNameList()
	if err != nil {
		WriteError(w, err)
		return
	}

	b, err := json.Marshal(names)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteResult(w, string(b))
	return
}

func nameRmHandler(w http.ResponseWriter, r *http.Request) {
	_, _, args, err := typeAuthArgs(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	name := args["name"]
	// name, addr := args["name"], args["addr"]
	// log.Debugf("name rm handler. name (%s). addr (%s)\n", name, addr)

	if name == "" {
		WriteError(w, fmt.Errorf("please specify a name"))
		return
	}

	if err := coreNameRm(name); err != nil {
		WriteError(w, err)
		return
	}

	WriteResult(w, fmt.Sprintf("Removed name (%s)", name))
}

// convenience function
func typeAuthArgs(r *http.Request) (typ string, auth string, args map[string]string, err error) {

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	// log.Debugln("Request body:", string(b))

	if err = json.Unmarshal(b, &args); err != nil {
		return
	}

	typ = args["type"]
	if typ == "" {
		typ = DefaultKeyType
	}

	auth = args["auth"]
	if auth == "" {
		auth = "" //DefaultAuth
	}

	return
}
