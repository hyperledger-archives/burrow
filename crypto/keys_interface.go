package crypto

type KeyClient interface {
	// Sign returns the signature bytes for given message signed with the key associated with signAddress
	Sign(signAddress Address, message []byte) (*Signature, error)

	// PublicKey returns the public key associated with a given address
	PublicKey(address Address) (publicKey PublicKey, err error)

	// Generate requests that a key be generate within the keys instance and returns the address
	Generate(keyName string, keyType CurveType) (keyAddress Address, err error)

	// Get the address for a keyname or the adress itself
	GetAddressForKeyName(keyName string) (keyAddress Address, err error)

	// Returns nil if the keys instance is healthy, error otherwise
	HealthCheck() error
}
