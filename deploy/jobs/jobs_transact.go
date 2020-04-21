package jobs

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/hyperledger/burrow/execution/exec"
	"github.com/hyperledger/burrow/logging"
	"github.com/hyperledger/burrow/txs/payload"
	"github.com/tmthrgd/go-hex"

	"github.com/hyperledger/burrow/deploy/def"
)

func FormulateSendJob(send *def.Send, account string, client *def.Client, logger *logging.Logger) (*payload.SendTx, error) {
	// Use Default
	send.Source = FirstOf(send.Source, account)

	// Formulate tx
	logger.InfoMsg("Sending Transaction",
		"source", send.Source,
		"destination", send.Destination,
		"amount", send.Amount)

	return client.Send(&def.SendArg{
		Input:    send.Source,
		Output:   send.Destination,
		Amount:   send.Amount,
		Sequence: send.Sequence,
	}, logger)
}

func SendJob(tx *payload.SendTx, client *def.Client, logger *logging.Logger) (string, error) {
	// Sign, broadcast, display
	txe, err := client.SignAndBroadcast(tx, logger)
	if err != nil {
		return "", fmt.Errorf("error in SendJob with payload %v: %w", tx, err)
	}

	LogTxExecution(txe, logger)
	if err != nil {
		return "", err
	}

	return txe.Receipt.TxHash.String(), nil
}

func FormulateRegisterNameJob(name *def.RegisterName, do *def.DeployArgs, playbook *def.Playbook, client *def.Client, logger *logging.Logger) ([]*payload.NameTx, error) {
	txs := make([]*payload.NameTx, 0)

	// If a data file is given it should be in csv format and
	// it will be read first. Once the file is parsed and sent
	// to the chain then a single nameRegTx will be sent if that
	// has been populated.
	if name.DataFile != "" {
		// open the file and use a reader
		var path string
		if filepath.IsAbs(name.DataFile) {
			path = name.Data
		} else {
			path = filepath.Join(playbook.Path, name.DataFile)
		}
		fileReader, err := os.Open(path)
		if err != nil {
			return nil, err
		}

		defer fileReader.Close()
		r := csv.NewReader(fileReader)

		// loop through the records
		for {
			// Read the record
			record, err := r.Read()

			// Catch the errors
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			// Sink the Amount into the third slot in the record if
			// it doesn't exist
			if len(record) <= 2 {
				record = append(record, name.Amount)
			}

			// Send an individual Tx for the record
			// [TODO]: move these to async using goroutines?
			r, err := registerNameTx(&def.RegisterName{
				Source:   name.Source,
				Name:     record[0],
				Data:     record[1],
				Amount:   record[2],
				Fee:      name.Fee,
				Sequence: name.Sequence,
			}, do, playbook.Account, client, logger)

			if err != nil {
				return nil, err
			}

			txs = append(txs, r)

			n := fmt.Sprintf("%s:%s", record[0], record[1])

			// TODO: write smarter
			if err = WriteJobResultCSV(n, r.String()); err != nil {
				return nil, err
			}
		}
	}

	// If the data field is populated then there is a single
	// nameRegTx to send. So do that *now*.
	if name.Data != "" {
		tx, err := registerNameTx(name, do, playbook.Account, client, logger)
		if err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}

	if len(txs) == 0 {
		return nil, fmt.Errorf("nothing to do")
	}

	return txs, nil
}

// Runs an individual nametx.
func registerNameTx(name *def.RegisterName, do *def.DeployArgs, account string, client *def.Client, logger *logging.Logger) (*payload.NameTx, error) {
	// Set Defaults
	name.Source = FirstOf(name.Source, account)
	name.Fee = FirstOf(name.Fee, do.DefaultFee)
	name.Amount = FirstOf(name.Amount, do.DefaultAmount)

	// Formulate tx
	logger.InfoMsg("NameReg Transaction",
		"name", name.Name,
		"data", name.Data,
		"amount", name.Amount)

	return client.Name(&def.NameArg{
		Input:    name.Source,
		Sequence: name.Sequence,
		Name:     name.Name,
		Amount:   name.Amount,
		Data:     name.Data,
		Fee:      name.Fee,
	}, logger)
}

func RegisterNameJob(txs []*payload.NameTx, client *def.Client, logger *logging.Logger) (string, error) {
	var result string

	for _, tx := range txs {
		// Sign, broadcast, display
		txe, err := client.SignAndBroadcast(tx, logger)
		if err != nil {
			return "", fmt.Errorf("error in RegisterNameJob with payload %v: %w", tx, err)
		}

		LogTxExecution(txe, logger)
		if err != nil {
			return "", err
		}

		result = txe.Receipt.TxHash.String()
	}
	return result, nil
}

func FormulatePermissionJob(perm *def.Permission, account string, client *def.Client, logger *logging.Logger) (*payload.PermsTx, error) {
	// Set defaults
	perm.Source = FirstOf(perm.Source, account)

	logger.TraceMsg("Permsision",
		"Target", perm.Target,
		"Marmots Deny", perm.Role,
		"Action", perm.Action)
	// Populate the transaction appropriately

	// Formulate tx
	return client.Permissions(&def.PermArg{
		Input:      perm.Source,
		Sequence:   perm.Sequence,
		Action:     perm.Action,
		Target:     perm.Target,
		Permission: perm.Permission,
		Role:       perm.Role,
		Value:      perm.Value,
	}, logger)
}

func PermissionJob(tx *payload.PermsTx, client *def.Client, logger *logging.Logger) (string, error) {
	logger.TraceMsg("Permissions returned in transaction: ", "args", tx.PermArgs)

	// Sign, broadcast, display
	txe, err := client.SignAndBroadcast(tx, logger)
	if err != nil {
		return "", fmt.Errorf("error in PermissionJob with payload %v: %w", tx, err)
	}

	LogTxExecution(txe, logger)
	if err != nil {
		return "", err
	}

	return txe.Receipt.TxHash.String(), nil
}

func FormulateIdentifyJob(id *def.Identify, account string, client *def.Client, logger *logging.Logger) (*payload.IdentifyTx, error) {
	// Use Default
	id.Source = FirstOf(id.Source, account)

	// Formulate tx
	logger.InfoMsg("Identify Transaction",
		"source", id.Source,
		"nodeKey", id.NodeKey,
		"netAddress", id.NetAddress)

	return client.Identify(&def.IdentifyArg{
		Input:      id.Source,
		Moniker:    id.Moniker,
		NodeKey:    id.NodeKey,
		NetAddress: id.NetAddress,
		Sequence:   id.Sequence,
	}, logger)
}

func IdentifyJob(tx *payload.IdentifyTx, client *def.Client, logger *logging.Logger) (string, error) {
	// Sign, broadcast, display
	txe, err := client.SignAndBroadcast(tx, logger)
	if err != nil {
		return "", fmt.Errorf("error in IdentifyJob with payload %v: %w", tx, err)
	}

	LogTxExecution(txe, logger)
	if err != nil {
		return "", err
	}

	return txe.Receipt.TxHash.String(), nil
}

func FirstOf(inputs ...string) string {
	for _, in := range inputs {
		if in != "" {
			return in
		}
	}
	return ""
}

func LogTxExecution(txe *exec.TxExecution, logger *logging.Logger) {
	// if there is nothing to unpack then just return.
	if txe == nil {
		return
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
}
