package util

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/hyperledger/burrow/deploy/compile"

	"github.com/hyperledger/burrow/execution/exec"
	log "github.com/sirupsen/logrus"
	"github.com/tmthrgd/go-hex"
)

// This is a closer function which is called by most of the tx_run functions
func ReadTxSignAndBroadcast(txe *exec.TxExecution, err error) error {
	// if there's an error just return.
	if err != nil {
		return err
	}

	// if there is nothing to unpack then just return.
	if txe == nil {
		return nil
	}

	// Unpack and display for the user.
	height := fmt.Sprintf("%d", txe.Height)

	if txe.Receipt.CreatesContract {
		log.WithField("addr", txe.Receipt.ContractAddress).Warn()
		log.WithField("txHash", txe.TxHash).Info()
	} else {
		log.WithField("=>", txe.TxHash).Warn("Transaction Hash")
		log.WithField("=>", height).Debug("Block height")
		ret := txe.GetResult().GetReturn()
		if len(ret) != 0 {
			log.WithField("=>", hex.EncodeUpperToString(ret)).Warn("Return Value")
			log.WithField("=>", txe.Exception).Debug("Exception")
		}
	}

	return nil
}

func ReadAbi(root, contract string) (string, error) {
	p := path.Join(root, stripHex(contract))
	if _, err := os.Stat(p); err != nil {
		log.WithField("abifile", p).Debug("Tried, not found")
		p = path.Join(root, stripHex(contract)+".bin")
		if _, err = os.Stat(p); err != nil {
			log.WithField("abifile", p).Debug("Tried, not found")
			return "", fmt.Errorf("Abi doesn't exist for =>\t%s", p)
		}
	}
	log.WithField("abifile", p).Debug("Found ABI")
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return "", err
	}
	sol := compile.SolidityOutputContract{}
	err = json.Unmarshal(b, &sol)
	if err != nil {
		return "", err
	}

	return string(sol.Abi), nil
}

func GetStringResponse(question string, defaultAnswer string, reader *os.File) (string, error) {
	readr := bufio.NewReader(reader)
	log.Warn(question)

	text, _ := readr.ReadString('\n')
	text = strings.Replace(text, "\n", "", 1)
	if text == "" {
		return defaultAnswer, nil
	}
	return text, nil
}

func GetIntResponse(question string, defaultAnswer int64, reader *os.File) (int64, error) {
	readr := bufio.NewReader(reader)
	log.Warn(question)

	text, _ := readr.ReadString('\n')
	text = strings.Replace(text, "\n", "", 1)
	if text == "" {
		return defaultAnswer, nil
	}

	result, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return 0, nil
	}

	return result, nil
}

// displays the question, scans for the response, if the response is an empty
// string will return default, otherwise will parseBool and return the result.
func GetBoolResponse(question string, defaultAnswer bool, reader *os.File) (bool, error) {
	var result bool
	readr := bufio.NewReader(reader)
	log.Warn(question)

	text, _ := readr.ReadString('\n')
	text = strings.Replace(text, "\n", "", 1)
	if text == "" {
		return defaultAnswer, nil
	}

	if text == "Yes" || text == "YES" || text == "Y" || text == "y" {
		result = true
	} else {
		result = false
	}

	return result, nil
}

func stripHex(s string) string {
	if len(s) > 1 {
		if s[:2] == "0x" {
			s = s[2:]
			if len(s)%2 != 0 {
				s = "0" + s
			}
			return s
		}
	}
	return s
}
