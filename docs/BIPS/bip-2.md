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
Burrow gives every node a validator key (even if not a validator).
Tendermint gives every peer a node key.
1. the node key is used (a lot) for the station-to-station protocol (kind of peer-to-peer TLS).
This effectively runs down the entropy in the key from the perspective of what you have revealed to a potential attacker.
Having a separate transport level key to your identity's signing key is 'good practice'.
1. is basically segregation - validator key may in principle have real-world value (validator voting power == bond) by keeping it used for the single purpose of signing votes the attack surface area (and frequency of signatures) is reduced.
Actually the node key doesn't get used for signatures, but it is still 'exposed' through the STS DH.

So we end up with a 1-to-1 correspondence but we have no way of mapping the two.

This is a reasonable think to give every node a validator key, and it is its primary identity, then a network-wide registry is necessary.

There is a lot of features/use cases where being able to lookup the p2p address (ID and NetAddress for that matter) will be useful, such as:
1. state channels/subnets.
1. Ops who often spent time trying to figure who's node is who's
1. Filter peers sync by node ID or address, allowing to forbid a peer to pull the chain state if it is not present in this registry

## Specification
<!--The technical specification should describe the syntax and semantics of any new feature.-->
The nodes submit their p2p identities by way of a handshake between the node private validator and the node p2p key.

The node broadcasts a transaction of a new type `IdentifyTx` signed by the validator key with the nodekey.

It also allows you to register and notify a replacement nodekey identity.

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
	// The node moniker name concatenated with the key id concatenated with '@' net address
    // signed by the node key signed by the validator key
	Signature []byte
}

type RegisteredNode {
    // Peer moniker name
	Moniker string
	// Validator node key
	NodeKey p2p.ID
	// Net address
	NetAddress string
}
```

A registry is available in the blockchain state, accessible by a getter method:
```go
func (s *State) GetNetworkRegistry() (map[crypto.Address][]*RegisteredNode, error)
```

A new route is available in node info:
`
GET /network/registry
`
Which returns:
```json
[
    {
        "address": "$VALIDATOR_ADDRESS",
        "pubKey":  "$VALIDATOR_PUB_KEY"
        "moniker": "$VALIDATOR_MONIKER",
        "nodeKey": "$VALIDATOR_NODE_KEY_ID",
        "netAddress": "$VALIDATOR_NODE_ADDRESS",
        "blockHeight": 2018
    }
]
```