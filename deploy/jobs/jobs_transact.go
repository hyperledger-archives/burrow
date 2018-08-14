package jobs

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/hyperledger/burrow/deploy/def"
	"github.com/hyperledger/burrow/deploy/util"
	log "github.com/sirupsen/logrus"
)

func SendJob(send *def.Send, do *def.Packages) (string, error) {
	// Use Default
	send.Source = useDefault(send.Source, do.Package.Account)

	// Formulate tx
	log.WithFields(log.Fields{
		"source":      send.Source,
		"destination": send.Destination,
		"amount":      send.Amount,
	}).Info("Sending Transaction")

	tx, err := do.Send(&def.SendArg{
		Input:    send.Source,
		Output:   send.Destination,
		Amount:   send.Amount,
		Sequence: send.Sequence,
	})
	if err != nil {
		return "", util.ChainErrorHandler(do, err)
	}

	// Sign, broadcast, display
	txe, err := do.SignAndBroadcast(tx)
	if err != nil {
		return "", util.ChainErrorHandler(do, err)
	}

	util.ReadTxSignAndBroadcast(txe, err)
	if err != nil {
		return "", err
	}

	return txe.Receipt.TxHash.String(), nil
}

func RegisterNameJob(name *def.RegisterName, do *def.Packages) (string, error) {
	// If a data file is given it should be in csv format and
	// it will be read first. Once the file is parsed and sent
	// to the chain then a single nameRegTx will be sent if that
	// has been populated.
	if name.DataFile != "" {
		// open the file and use a reader
		fileReader, err := os.Open(name.DataFile)
		if err != nil {
			return "", err
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
				return "", err
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
			}, do)

			if err != nil {
				return "", err
			}

			n := fmt.Sprintf("%s:%s", record[0], record[1])

			// TODO: write smarter
			if err = WriteJobResultCSV(n, r); err != nil {
				return "", err
			}
		}
	}

	// If the data field is populated then there is a single
	// nameRegTx to send. So do that *now*.
	if name.Data != "" {
		return registerNameTx(name, do)
	} else {
		return "data_file_parsed", nil
	}
}

// Runs an individual nametx.
func registerNameTx(name *def.RegisterName, do *def.Packages) (string, error) {
	// Set Defaults
	name.Source = useDefault(name.Source, do.Package.Account)
	name.Fee = useDefault(name.Fee, do.DefaultFee)
	name.Amount = useDefault(name.Amount, do.DefaultAmount)

	// Formulate tx
	log.WithFields(log.Fields{
		"name":   name.Name,
		"data":   name.Data,
		"amount": name.Amount,
	}).Info("NameReg Transaction")

	tx, err := do.Name(&def.NameArg{
		Input:    name.Source,
		Sequence: name.Sequence,
		Name:     name.Name,
		Amount:   name.Amount,
		Data:     name.Data,
		Fee:      name.Fee,
	})
	if err != nil {
		return "", util.ChainErrorHandler(do, err)
	}
	// Sign, broadcast, display
	txe, err := do.SignAndBroadcast(tx)
	if err != nil {
		return "", util.ChainErrorHandler(do, err)
	}

	util.ReadTxSignAndBroadcast(txe, err)
	if err != nil {
		return "", err
	}

	return txe.Receipt.TxHash.String(), nil
}

func PermissionJob(perm *def.Permission, do *def.Packages) (string, error) {
	// Set defaults
	perm.Source = useDefault(perm.Source, do.Package.Account)

	log.Debug("Target: ", perm.Target)
	log.Debug("Marmots Deny: ", perm.Role)
	log.Debug("Action: ", perm.Action)
	// Populate the transaction appropriately

	// Formulate tx
	tx, err := do.Permissions(&def.PermArg{
		Input:      perm.Source,
		Sequence:   perm.Sequence,
		Action:     perm.Action,
		Target:     perm.Target,
		Permission: perm.Permission,
		Role:       perm.Role,
		Value:      perm.Value,
	})
	if err != nil {
		return "", util.ChainErrorHandler(do, err)
	}

	log.Debug("What are the args returned in transaction: ", tx.PermArgs)

	// Sign, broadcast, display
	txe, err := do.SignAndBroadcast(tx)
	if err != nil {
		return "", util.ChainErrorHandler(do, err)
	}

	util.ReadTxSignAndBroadcast(txe, err)
	if err != nil {
		return "", err
	}

	return txe.Receipt.TxHash.String(), nil
}

func useDefault(thisOne, defaultOne string) string {
	if thisOne == "" {
		return defaultOne
	}
	return thisOne
}
