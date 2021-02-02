package crypto

/*
	Burrow supports ed25519 and secp256k1 signing. The *Signature type stores the signature scheme under CurveType
	and the raw bytes of each signature type are stored in Signature.Signature.

	For secp256k1 we use the btcec compact representation including magic parity byte prefix: byte(27) or byte(28) to
	represent the 0 or 1 parity for the y-coordinate of the public key.

*/
