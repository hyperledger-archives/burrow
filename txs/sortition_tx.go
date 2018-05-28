package txs

import (
	"fmt"
	"io"

	acm "github.com/hyperledger/burrow/account"
	"github.com/tendermint/go-wire"
)

type SortitionTx struct {
	PublicKey acm.PublicKey `json:"public_key"`
	Signature acm.Signature `json:"signature"`
	Height    uint64        `json:"height"`
	Index     uint64        `json:"index"`
	Proof     []byte        `json:"proof"`
	txHashMemoizer
}

var _ Tx = &SortitionTx{}

func NewSortitionTx(publicKey acm.PublicKey, height, index uint64, proof []byte) *SortitionTx {
	return &SortitionTx{
		PublicKey: publicKey,
		Height:    height,
		Index:     index,
		Proof:     proof,
	}
}

func (tx *SortitionTx) WriteSignBytes(chainID string, w io.Writer, n *int, err *error) {

	jsonTx := fmt.Sprintf(`{"public_key":%v,"height":"%v","index":%v,"proof":%v}`, tx.PublicKey, tx.Height, tx.Index, tx.Proof)

	wire.WriteTo([]byte(fmt.Sprintf(`{"chain_id":%s,"tx":[%v,%v]}`,
		jsonEscape(chainID),
		TxTypeSortition,
		jsonTx)), w, n, err)
}

func (tx *SortitionTx) GetInputs() []TxInput {
	return nil
}

func (tx *SortitionTx) String() string {
	return fmt.Sprintf("SortitionTx{address:%v, height: %v, index: %v}", tx.PublicKey.Address(), tx.Height, tx.Index)
}

func (tx *SortitionTx) Hash(chainID string) []byte {
	return tx.txHashMemoizer.hash(chainID, tx)
}

func (tx *SortitionTx) Sign(chainID string, signingAccounts ...acm.AddressableSigner) error {
	if len(signingAccounts) != 1 {
		return fmt.Errorf("SortitionTx expects a single AddressableSigner for its single Input but %v were provieded",
			len(signingAccounts))
	}
	var err error
	tx.PublicKey = signingAccounts[0].PublicKey()
	tx.Signature, err = acm.ChainSign(signingAccounts[0], chainID, tx)
	if err != nil {
		return fmt.Errorf("could not sign %v: %v", tx, err)
	}
	return nil
}
