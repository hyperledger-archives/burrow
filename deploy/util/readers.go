package util

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging"
	hex "github.com/tmthrgd/go-hex"
)

// This is a closer function which is called by most of the tx_run functions
func ReadTxSignAndBroadcast(txe *exec.TxExecution, err error, logger *logging.Logger) error {
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
		logger.InfoMsg("Tx Return",
			"addr", txe.Receipt.ContractAddress.String(),
			"Transaction Hash", hex.EncodeToString(txe.TxHash))
	} else {
		logger.InfoMsg("Tx Return",
			"Transaction Hash", hex.EncodeToString(txe.TxHash),
			"Block Height", height)

		ret := txe.GetResult().GetReturn()
		if len(ret) != 0 {
			logger.InfoMsg("Return",
				"Return Value", hex.EncodeUpperToString(ret),
				"Exception", txe.Exception)
		}
	}

	return nil
}

func GetStringResponse(question string, defaultAnswer string, reader *os.File, logger *logging.Logger) (string, error) {
	readr := bufio.NewReader(reader)
	logger.InfoMsg(question)

	text, _ := readr.ReadString('\n')
	text = strings.Replace(text, "\n", "", 1)
	if text == "" {
		return defaultAnswer, nil
	}
	return text, nil
}

func GetIntResponse(question string, defaultAnswer int64, reader *os.File, logger *logging.Logger) (int64, error) {
	readr := bufio.NewReader(reader)

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
func GetBoolResponse(question string, defaultAnswer bool, reader *os.File, logger *logging.Logger) (bool, error) {
	var result bool
	readr := bufio.NewReader(reader)
	logger.InfoMsg(question)

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
