package txs

import (
	"fmt"
	"io"

	acm "github.com/hyperledger/burrow/account"
	"github.com/hyperledger/burrow/account/state"
	"github.com/tendermint/go-wire"
)

type BondTx struct {
	PubKey    acm.PublicKey
	Signature acm.Signature
	Inputs    []*TxInput
	UnbondTo  []*TxOutput
	txHashMemoizer
}

var _ Tx = &BondTx{}

func NewBondTx(pubkey acm.PublicKey) (*BondTx, error) {
	return &BondTx{
		PubKey:   pubkey,
		Inputs:   []*TxInput{},
		UnbondTo: []*TxOutput{},
	}, nil
}

func (tx *BondTx) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {
	wire.WriteTo([]byte(fmt.Sprintf(`{"chain_id":%s`, jsonEscape(chainID))), w, n, err)
	wire.WriteTo([]byte(fmt.Sprintf(`,"tx":[%v,{"inputs":[`, TxTypeBond)), w, n, err)
	for i, in := range tx.Inputs {
		in.WriteSignBytes(w, n, err)
		if i != len(tx.Inputs)-1 {
			wire.WriteTo([]byte(","), w, n, err)
		}
	}
	wire.WriteTo([]byte(fmt.Sprintf(`],"pub_key":`)), w, n, err)
	wire.WriteTo(wire.JSONBytes(tx.PubKey), w, n, err)
	wire.WriteTo([]byte(`,"unbond_to":[`), w, n, err)
	for i, out := range tx.UnbondTo {
		out.WriteSignBytes(w, n, err)
		if i != len(tx.UnbondTo)-1 {
			wire.WriteTo([]byte(","), w, n, err)
		}
	}
	wire.WriteTo([]byte(`]}]}`), w, n, err)
}

func (tx *BondTx) GetInputs() []TxInput {
	return copyInputs(tx.Inputs)
}

func (tx *BondTx) String() string {
	return fmt.Sprintf("BondTx{%v: %v -> %v}", tx.PubKey, tx.Inputs, tx.UnbondTo)
}

func (tx *BondTx) Hash(chainID string) []byte {
	return tx.txHashMemoizer.hash(chainID, tx)
}

func (tx *BondTx) AddInput(st state.AccountGetter, pubkey acm.PublicKey, amt uint64) error {
	addr := pubkey.Address()
	acc, err := st.GetAccount(addr)
	if err != nil {
		return err
	}
	if acc == nil {
		return fmt.Errorf("Invalid address %s from pubkey %s", addr, pubkey)
	}
	return tx.AddInputWithSequence(pubkey, amt, acc.Sequence()+uint64(1))
}

func (tx *BondTx) AddInputWithSequence(pubkey acm.PublicKey, amt uint64, sequence uint64) error {
	tx.Inputs = append(tx.Inputs, &TxInput{
		Address:   pubkey.Address(),
		Amount:    amt,
		Sequence:  sequence,
		PublicKey: pubkey,
	})
	return nil
}

func (tx *BondTx) AddOutput(addr acm.Address, amt uint64) error {
	tx.UnbondTo = append(tx.UnbondTo, &TxOutput{
		Address: addr,
		Amount:  amt,
	})
	return nil
}

func (tx *BondTx) Sign(chainID string, signingAccounts ...acm.AddressableSigner) error {
	if len(signingAccounts) != len(tx.Inputs)+1 {
		return fmt.Errorf("BondTx expects %v SigningAccounts but got %v", len(tx.Inputs)+1,
			len(signingAccounts))
	}
	var err error
	tx.Signature, err = acm.ChainSign(signingAccounts[0], chainID, tx)
	if err != nil {
		return fmt.Errorf("could not sign %v: %v", tx, err)
	}
	for i := 1; i <= len(signingAccounts); i++ {
		tx.Inputs[i].PublicKey = signingAccounts[i].PublicKey()
		tx.Inputs[i].Signature, err = acm.ChainSign(signingAccounts[i], chainID, tx)
		if err != nil {
			return fmt.Errorf("could not sign tx %v input %v: %v", tx, tx.Inputs[i], err)
		}
	}
	return nil
}
