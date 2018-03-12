package inmemory

import (
	"crypto/rand"
	"errors"
	"fmt"
	acm "github.com/hyperledger/burrow/account"
	"github.com/tendermint/ed25519"
	"github.com/tendermint/go-crypto"
)

type signer struct {
	conf *InMemoryCrypto
}

func Signer(conf *InMemoryCrypto) (acm.Signer, error) {
	localSigner := &signer{
		conf: conf,
	}

	msg := []byte("SignIt")

	sign, err := localSigner.Sign(msg)
	if err != nil {
		return nil, fmt.Errorf("Unable to sign with provided private key: %v", err)
	}

	var signatureEd25519 [ed25519.SignatureSize]byte
	copy(signatureEd25519[:], sign.Bytes()[1:])
	if !localSigner.Verify(&conf.PublicKey, msg, &signatureEd25519) {
		return nil, errors.New("Unable to verify own signed message with provided keys pair")
	}

	return localSigner, nil
}

func (signer *signer) Sign(msg []byte) (acm.Signature, error) {
	var privKeyEd25519 [ed25519.PrivateKeySize]byte
	copy(privKeyEd25519[:], signer.conf.PrivateKey.Bytes()[1:])
	return acm.SignatureFromBytes(ed25519.Sign(&privKeyEd25519, msg)[:])
}

func (signer *signer) Verify(publicKey *acm.PublicKey, message []byte, sig *[ed25519.SignatureSize]byte) bool {
	var pubKeyEd25519 [ed25519.PublicKeySize]byte
	copy(pubKeyEd25519[:], publicKey.Bytes()[1:])
	return ed25519.Verify(&pubKeyEd25519, message, sig)
}

func GenPrivateKey() (privateKey *[ed25519.PrivateKeySize]byte, err error) {
	_, privateKey, err = ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("Error while generating key %v\n", err)
	}

	return
}

type keyAddressable struct {
	publicKey acm.PublicKey
	address   acm.Address
}

func (ka *keyAddressable) Address() acm.Address {
	return ka.address
}

func (ka *keyAddressable) PublicKey() acm.PublicKey {
	return ka.publicKey
}

func Addressable(conf *InMemoryCrypto) (acm.Addressable, error) {
	var pkBytes [ed25519.PublicKeySize]byte
	copy(pkBytes[:], conf.PublicKey.Bytes()[1:])
	pubKeyEd25519 := crypto.PubKeyEd25519(pkBytes)

	return &keyAddressable{
		address:   acm.MustAddressFromBytes(pubKeyEd25519.Address()),
		publicKey: conf.PublicKey,
	}, nil
}
