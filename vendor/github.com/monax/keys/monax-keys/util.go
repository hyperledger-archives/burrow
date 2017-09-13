package keys

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/monax/keys/common"
)

//------------------------------------------------------------
// auth

func hiddenAuth() string {
	fmt.Printf("Enter Password:")
	pwd, err := gopass.GetPasswdMasked()
	if err != nil {
		common.IfExit(err)
	}
	return string(pwd)
}

//------------------------------------------------------------
// key names

// most commands require at least one of --name or --addr
func checkGetNameAddr(name, addr string) string {
	addr, err := getNameAddr(name, addr)
	common.IfExit(err)
	return addr
}

// return addr from name or addr
func getNameAddr(name, addr string) (string, error) {
	if name == "" && addr == "" {
		return "", fmt.Errorf("at least one of --name or --addr must be provided")
	}

	// name takes precedent if both are given
	var err error
	if name != "" {
		addr, err = coreNameGet(name)
		if err != nil {
			return "", err
		}
	}
	return strings.ToUpper(addr), nil
}

//------------------------------------------------------------
// http client

func unpackResponse(resp *http.Response) (string, string, error) {
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	r := new(HTTPResponse)
	if err := json.Unmarshal(b, r); err != nil {
		return "", "", fmt.Errorf("Error unmarshaling response: %v", err)
	}
	return r.Response, r.Error, nil
}

type ErrConnectionRefused string

func (e ErrConnectionRefused) Error() string {
	return string(e)
}

func requestResponse(req *http.Request) (string, string, error) {
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return "", "", ErrConnectionRefused(err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", "", fmt.Errorf(resp.Status)
	}
	return unpackResponse(resp)
}

// Call the http server
func Call(method string, args map[string]string) (string, error) {
	url := fmt.Sprintf("%s/%s", DaemonAddr, method)
	b, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("Error marshaling args map: %v", err)
	}
	// log.Debugln("calling", url)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(b))
	r, errS, err := requestResponse(req)
	if err != nil {
		return "", err
	}
	if errS != "" {
		return "", fmt.Errorf(errS)
	}
	return r, nil
}
