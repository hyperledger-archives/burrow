---
bip: 2
title: Identify Transaction
author: Casey Kuhlman (@compleatang), Silas Davis (@silasdavis), Pierrick Hymbert (@phymbert)
discussions-to: https://chat.hyperledger.org/channel/burrow-contributors
status: Draft
type: Standards Track
category: State, Consensus, Governance, Gateway
created: 2018-09-10
---

## Node keys and validators registry
This BIP introduces Identify Transaction to register and track node key related to their validator key.

## Motivation

There is 2 different kind of keys in a running Burrow node:
- ABCI validator key (even if not a block validator)
- P2P node key

This is basically segregation:

1. The node key is used (a lot) for the station-to-station (STS) protocol (kind of peer-to-peer TLS).
This effectively runs down the entropy in the key from the perspective of what you have revealed to a potential attacker.
Having a separate transport level key to your identity's signing key is 'good practice'.
1. Validator key may in principle have real-world value (validator voting power == bond) by keeping it used for the single purpose of signing votes the attack surface area (and frequency of signatures) is reduced.
Actually the node key doesn't get used for signatures, but it is still 'exposed' through the STS DH.

So we end up with a 1-to-1 key correspondence but we have no way of mapping the two.

This is a reasonable think to give every node a validator key, and it is its primary identity, then a network-wide registry is necessary.

There is a lot of features/use cases where being able to lookup the p2p address (ID and NetAddress for that matter) will be useful, such as:
1. state channels/subnets.
1. Ops who often spent time trying to figure who's node is who's
1. Filter peers sync by node ID or address, allowing to forbid a peer to pull the chain state if it is not present in this registry

## Specification
Nodes submit their p2p identities by way of a handshake between the node private validator and the node p2p key.

The node broadcasts a transaction of a new type `IdentifyTx` signed by the validator key with the node key.

It also allows to register and notify a replacement node key identity.

Burrow verifies a multisig of this tx of two inputs: validator key, node key.

If they mutually sign then that key mapping gets added to network-wide registry, a simple store.

A new transaction type is available:
```go
type IdentifyTx struct {
    // Sender
	Input *TxInput
	// Validator address
	Address crypto.Address
	// Validator public key
	PubKey crypto.PublicKey
	// Node
	Node *RegisteredNode
	// The RegisteredNode.String() multisigned by the node key and the validator key
	Signature []byte
}

type RegisteredNode {
    // Peer moniker name
	Moniker string
	// Node key id (crypto address)
	ID p2p.ID
	// Node key public key
	PublicKey crypto.PublicKey
	// Net address
	NetAddress string
}
```

A registry is available in the blockchain state, accessible by a getter method:
```go
// GetNetworkRegistry returns for each validator address, the list of their identified node at the current state
func (s *State) GetNetworkRegistry() (map[crypto.Address][]*RegisteredNode, error)
```

A new route is available in node info:
`
GET /network/registry
`

Which returns:
```javascript
[
    {
        "address": "$VALIDATOR_ADDRESS",
        "pubKey":  "$VALIDATOR_PUB_KEY",
        "moniker": "$VALIDATOR_MONIKER",
        "nodeKey": "$VALIDATOR_NODE_KEY_ID",
        "netAddress": "$VALIDATOR_NODE_ADDRESS"
    }
]
```