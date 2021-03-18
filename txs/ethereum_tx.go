package txs

import (
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/encoding"
	"github.com/hyperledger/burrow/encoding/rlp"
	"github.com/tmthrgd/go-hex"
)

// Order matters for serialisation
type EthRawTx struct {
	Sequence uint64   `json:"nonce"`
	GasPrice uint64   `json:"gasPrice"`
	GasLimit uint64   `json:"gasLimit"`
	To       []byte   `json:"to"`
	Amount   *big.Int `json:"value"`
	Data     []byte   `json:"data"`
	// Signature
	V *big.Int
	R *big.Int
	S *big.Int
	// Included in hash but not part of serialised message
	chainID *big.Int
}

func NewEthRawTx(chainID *big.Int) *EthRawTx {
	return &EthRawTx{chainID: chainID}
}

func EthRawTxFromEnvelope(txEnv *Envelope) (*EthRawTx, error) {
	if txEnv.GetEncoding() != Envelope_RLP {
		return nil, fmt.Errorf("can only form EthRawTx from RLP-encoded Envelope")
	}
	rawTx, err := txEnv.Tx.RLPRawTx()
	if err != nil {
		return nil, err
	}
	if len(txEnv.Signatories) == 0 {
		return rawTx, nil
	}
	if len(txEnv.Signatories) > 1 {
		return nil, fmt.Errorf("can only form EthRawTx from Envelope with a zero or one signatories")
	}
	sig, err := txEnv.Signatories[0].Signature.GetEthSignature(encoding.GetEthChainID(txEnv.Tx.ChainID))
	if err != nil {
		return nil, err
	}
	// Link signature values into EthRawTx
	rawTx.V = &sig.V
	rawTx.R = &sig.R
	rawTx.S = &sig.S
	return rawTx, nil
}

func (tx *EthRawTx) RecoverPublicKey() (*crypto.PublicKey, *crypto.Signature, error) {
	hash, err := tx.Hash()
	if err != nil {
		return nil, nil, err
	}
	if tx.R.Sign() == 0 || tx.S.Sign() == 0 {
		return nil, nil, fmt.Errorf("EthRawTx does not appear to be signed")
	}
	ethSig := crypto.EIP155Signature{
		Secp256k1Signature: crypto.Secp256k1Signature{
			V: *tx.V,
			R: *tx.R,
			S: *tx.S,
		},
	}
	compactSig, err := ethSig.ToCompactSignature()
	if err != nil {
		return nil, nil, err
	}
	pubKey, compressed, err := btcec.RecoverCompact(btcec.S256(), compactSig, hash)
	if err != nil {
		return nil, nil, err
	}
	var pubKeyBytes []byte
	if compressed {
		pubKeyBytes = pubKey.SerializeCompressed()
	} else {
		pubKeyBytes = pubKey.SerializeUncompressed()
	}
	publicKey, err := crypto.PublicKeyFromBytes(pubKeyBytes, crypto.CurveTypeSecp256k1)
	if err != nil {
		return nil, nil, err
	}
	signature, err := crypto.SignatureFromBytes(compactSig, crypto.CurveTypeSecp256k1)
	if err != nil {
		return nil, nil, err
	}
	return publicKey, signature, nil
}

func (tx *EthRawTx) SignBytes() ([]byte, error) {
	return rlp.Encode([]interface{}{
		tx.Sequence,
		tx.GasPrice,
		tx.GasLimit,
		tx.To,
		tx.Amount,
		tx.Data,
		tx.chainID,
		uint(0), uint(0),
	})
}

func (tx *EthRawTx) Hash() ([]byte, error) {
	enc, err := tx.SignBytes()
	if err != nil {
		return nil, err
	}
	return crypto.Keccak256(enc), nil
}

func (tx *EthRawTx) Marshal() ([]byte, error) {
	return rlp.Encode(tx)
}

func (tx *EthRawTx) MarshalString() (string, error) {
	bs, err := tx.Marshal()
	if err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(bs), nil
}
