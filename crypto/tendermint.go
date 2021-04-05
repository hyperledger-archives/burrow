package crypto

import (
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	tmCrypto "github.com/tendermint/tendermint/crypto"
	tmEd25519 "github.com/tendermint/tendermint/crypto/ed25519"
	tmSecp256k1 "github.com/tendermint/tendermint/crypto/secp256k1"
)

func PublicKeyFromTendermintPubKey(pubKey tmCrypto.PubKey) (*PublicKey, error) {
	switch pk := pubKey.(type) {
	case tmEd25519.PubKey:
		return PublicKeyFromBytes(pk[:], CurveTypeEd25519)
	case tmSecp256k1.PubKey:
		return PublicKeyFromBytes(pk[:], CurveTypeSecp256k1)
	default:
		return nil, fmt.Errorf("unrecognised tendermint public key type: %v", pk)
	}
}

// PublicKey extensions

func (p PublicKey) TendermintPubKey() tmCrypto.PubKey {
	switch p.CurveType {
	case CurveTypeEd25519:
		return tmEd25519.PubKey(p.PublicKey)
	case CurveTypeSecp256k1:
		return tmSecp256k1.PubKey(p.PublicKey)
	default:
		return nil
	}
}

func (p PublicKey) TendermintAddress() tmCrypto.Address {
	switch p.CurveType {
	case CurveTypeEd25519:
		return tmCrypto.Address(p.GetAddress().Bytes())
	case CurveTypeSecp256k1:
		// Tendermint represents addresses like Bitcoin
		return tmCrypto.Address(RIPEMD160(SHA256(p.PublicKey[:])))
	default:
		panic(fmt.Sprintf("unknown CurveType %d", p.CurveType))
	}
}

// Signature extensions

func (sig Signature) TendermintSignature() []byte {
	switch sig.CurveType {
	case CurveTypeSecp256k1:
		sig, err := btcec.ParseDERSignature(sig.GetSignature(), btcec.S256())
		if err != nil {
			return nil
		}
		// https://github.com/tendermint/tendermint/blob/master/crypto/secp256k1/secp256k1_nocgo.go#L62
		r := sig.R.Bytes()
		s := sig.S.Bytes()
		data := make([]byte, 64)
		copy(data[32-len(r):32], r)
		copy(data[64-len(s):64], s)
		return data
	}
	return sig.Signature
}
