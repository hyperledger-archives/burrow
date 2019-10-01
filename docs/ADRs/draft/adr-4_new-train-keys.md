---
adr: 4
title: Securing validator and participants keys
author: Silas Davis (@silasdavis), Sean Young (@seanyoung), Greg Hill (@gregdhill)
discussions-to: https://chat.hyperledger.org/channel/burrow-contributors
status: Draft
type: Non-Standards Track
category: Security, Keys management
created: 2019-10-01
---

## Securing validator and participants keys

This ADR introduces how we will ensure that private key matter is protected.

## Motivation

A burrow validating node can serves two purposes; one is to validate blocks and contribute to concensus and
block production. The other function is to receive transactions, and add them to the blocks.

Those transactions are unsigned, and are signed by the burrow node if it has its private key available. This
is called mempool signing. So, this port (the GRPC port) should not be exposed to the internet.

The burrow server assumes that different clients that have access to the GRPC port can trust each other. This
is so because all the keys which are accessible to burrow can be used for signing, without authentication.
This not always desirable.

In addition, the keys server mixes validator keys with participants keys, so it is possible to sign a transaction
using the validator key.

A burrow network is blockchain; as such we expect authenticity but do not expect confidentiality.

Furthermore encryption is not a goal in it-self, however the private key matter must be protected, so there
should be a tls and basic auth available for the keys server.

## Node key is stored in keystore

The node key is now stored in the keystore. This is done so we can sign the IdentifyTx response with both
the validator key and the node key.

## burrow keys service to be replaced by burrow proxy

The keys service should not offer a generic "sign" method. Rather, it should only sign transactions which
are destined for the chain. In order to do this, it needs to know the sequence for the account. The sequence
number increases for successful transaction, so the keys service will need to connect to the burrow validator
and submit the transaction for the user.

So the keys service is now called the proxy service, and also proxies all the other services; rpcevent, rpcquery
and rpcinfo. It also offers the keys service (minus the sign functionality).

Any number of proxy servers can be run, and each with only the keys they need. In addition, the burrow server
can run with only the keys it needs (node key and validator key)

## burrow keys and burrow deploy cli should work with keys directory

burrow deploy can now do local signing with access to the key store.

## burrow.js

burrow js should work via the proxy server, since it cannot do local signing.

## Encryption and authentication to the key service (TLS and basic auth)

The keys service should provide a mode where it is available in via an encrypted and authenicated setup.
This is not required for our own infrastructure.
