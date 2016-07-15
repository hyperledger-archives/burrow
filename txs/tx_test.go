package txs

import (
	"testing"

	acm "github.com/eris-ltd/eris-db/account"
	ptypes "github.com/eris-ltd/eris-db/permission/types"

	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-crypto"
	//"github.com/tendermint/tendermint/types"
)

var chainID = "myChainID"

func TestSendTxSignable(t *testing.T) {
	sendTx := &SendTx{
		Inputs: []*TxInput{
			&TxInput{
				Address:  []byte("input1"),
				Amount:   12345,
				Sequence: 67890,
			},
			&TxInput{
				Address:  []byte("input2"),
				Amount:   111,
				Sequence: 222,
			},
		},
		Outputs: []*TxOutput{
			&TxOutput{
				Address: []byte("output1"),
				Amount:  333,
			},
			&TxOutput{
				Address: []byte("output2"),
				Amount:  444,
			},
		},
	}
	signBytes, err := acm.SignBytes(chainID, bondTx)
	if err != nil {
		t.Fatal(err)
	}
	signStr := string(signBytes)
	expected := Fmt(`{"chain_id":"%s","tx":[1,{"inputs":[{"address":"696E70757431","amount":12345,"sequence":67890},{"address":"696E70757432","amount":111,"sequence":222}],"outputs":[{"address":"6F757470757431","amount":333},{"address":"6F757470757432","amount":444}]}]}`,
		chainID)

	if signStr != expected {
		t.Errorf("Got unexpected sign string for SendTx. Expected:\n%v\nGot:\n%v", expected, signStr)
	}
}

func TestCallTxSignable(t *testing.T) {
	callTx := &CallTx{
		Input: &TxInput{
			Address:  []byte("input1"),
			Amount:   12345,
			Sequence: 67890,
		},
		Address:  []byte("contract1"),
		GasLimit: 111,
		Fee:      222,
		Data:     []byte("data1"),
	}
	signBytes, err := acm.SignBytes(chainID, bondTx)
	if err != nil {
		t.Fatal(err)
	}
	signStr := string(signBytes)
	expected := Fmt(`{"chain_id":"%s","tx":[2,{"address":"636F6E747261637431","data":"6461746131","fee":222,"gas_limit":111,"input":{"address":"696E70757431","amount":12345,"sequence":67890}}]}`,
		chainID)
	if signStr != expected {
		t.Errorf("Got unexpected sign string for CallTx. Expected:\n%v\nGot:\n%v", expected, signStr)
	}
}

func TestNameTxSignable(t *testing.T) {
	nameTx := &NameTx{
		Input: &TxInput{
			Address:  []byte("input1"),
			Amount:   12345,
			Sequence: 250,
		},
		Name: "google.com",
		Data: "secretly.not.google.com",
		Fee:  1000,
	}
	signBytes, err := acm.SignBytes(chainID, bondTx)
	if err != nil {
		t.Fatal(err)
	}
	signStr := string(signBytes)
	expected := Fmt(`{"chain_id":"%s","tx":[3,{"data":"secretly.not.google.com","fee":1000,"input":{"address":"696E70757431","amount":12345,"sequence":250},"name":"google.com"}]}`,
		chainID)
	if signStr != expected {
		t.Errorf("Got unexpected sign string for CallTx. Expected:\n%v\nGot:\n%v", expected, signStr)
	}
}

func TestBondTxSignable(t *testing.T) {
	privKeyBytes := make([]byte, 64)
	privAccount := acm.GenPrivAccountFromPrivKeyBytes(privKeyBytes)
	bondTx := &BondTx{
		PubKey: privAccount.PubKey.(crypto.PubKeyEd25519),
		Inputs: []*TxInput{
			&TxInput{
				Address:  []byte("input1"),
				Amount:   12345,
				Sequence: 67890,
			},
			&TxInput{
				Address:  []byte("input2"),
				Amount:   111,
				Sequence: 222,
			},
		},
		UnbondTo: []*TxOutput{
			&TxOutput{
				Address: []byte("output1"),
				Amount:  333,
			},
			&TxOutput{
				Address: []byte("output2"),
				Amount:  444,
			},
		},
	}
	signBytes, err := acm.SignBytes(chainID, bondTx)
	if err != nil {
		t.Fatal(err)
	}
	signStr := string(signBytes)
	expected := Fmt(`{"chain_id":"%s","tx":[17,{"inputs":[{"address":"696E70757431","amount":12345,"sequence":67890},{"address":"696E70757432","amount":111,"sequence":222}],"pub_key":"3B6A27BCCEB6A42D62A3A8D02A6F0D73653215771DE243A63AC048A18B59DA29","unbond_to":[{"address":"6F757470757431","amount":333},{"address":"6F757470757432","amount":444}]}]}`,
		chainID)
	if signStr != expected {
		t.Errorf("Unexpected sign string for BondTx. \nGot %s\nExpected %s", signStr, expected)
	}
}

func TestUnbondTxSignable(t *testing.T) {
	unbondTx := &UnbondTx{
		Address: []byte("address1"),
		Height:  111,
	}
	signBytes, err := acm.SignBytes(chainID, bondTx)
	if err != nil {
		t.Fatal(err)
	}
	signStr := string(signBytes)
	expected := Fmt(`{"chain_id":"%s","tx":[18,{"address":"6164647265737331","height":111}]}`,
		chainID)
	if signStr != expected {
		t.Errorf("Got unexpected sign string for UnbondTx")
	}
}

func TestRebondTxSignable(t *testing.T) {
	rebondTx := &RebondTx{
		Address: []byte("address1"),
		Height:  111,
	}
	signBytes, err := acm.SignBytes(chainID, bondTx)
	if err != nil {
		t.Fatal(err)
	}
	signStr := string(signBytes)
	expected := Fmt(`{"chain_id":"%s","tx":[19,{"address":"6164647265737331","height":111}]}`,
		chainID)
	if signStr != expected {
		t.Errorf("Got unexpected sign string for RebondTx")
	}
}

func TestPermissionsTxSignable(t *testing.T) {
	permsTx := &PermissionsTx{
		Input: &TxInput{
			Address:  []byte("input1"),
			Amount:   12345,
			Sequence: 250,
		},
		PermArgs: &ptypes.SetBaseArgs{
			Address:    []byte("address1"),
			Permission: 1,
			Value:      true,
		},
	}

	signBytes, err := acm.SignBytes(chainID, bondTx)
	if err != nil {
		t.Fatal(err)
	}
	signStr := string(signBytes)
	expected := Fmt(`{"chain_id":"%s","tx":[32,{"args":"[2,{"address":"6164647265737331","permission":1,"value":true}]","input":{"address":"696E70757431","amount":12345,"sequence":250}}]}`,
		chainID)
	if signStr != expected {
		t.Errorf("Got unexpected sign string for PermsTx. Expected:\n%v\nGot:\n%v", expected, signStr)
	}
}

/*
func TestDupeoutTxSignable(t *testing.T) {
	privAcc := acm.GenPrivAccount()
	partSetHeader := types.PartSetHeader{Total: 10, Hash: []byte("partsethash")}
	voteA := &types.Vote{
		Height:           10,
		Round:            2,
		Type:             types.VoteTypePrevote,
		BlockHash:        []byte("myblockhash"),
		BlockPartsHeader: partSetHeader,
	}
	sig := privAcc.Sign(chainID, voteA)
	voteA.Signature = sig.(crypto.SignatureEd25519)
	voteB := voteA.Copy()
	voteB.BlockHash = []byte("myotherblockhash")
	sig = privAcc.Sign(chainID, voteB)
	voteB.Signature = sig.(crypto.SignatureEd25519)

	dupeoutTx := &DupeoutTx{
		Address: []byte("address1"),
		VoteA:   *voteA,
		VoteB:   *voteB,
	}
	signBytes := acm.SignBytes(chainID, dupeoutTx)
	signStr := string(signBytes)
	expected := Fmt(`{"chain_id":"%s","tx":[20,{"address":"6164647265737331","vote_a":%v,"vote_b":%v}]}`,
		chainID, *voteA, *voteB)
	if signStr != expected {
		t.Errorf("Got unexpected sign string for DupeoutTx")
	}
}*/
