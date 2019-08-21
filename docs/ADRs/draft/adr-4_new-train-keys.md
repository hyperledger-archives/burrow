---
adr: 4
title: Securing validator and participants keys
author: Silas Davis (@silasdavis), Sean Young (@seanyoung), Greg Hill (@gregdhill)
discussions-to: https://chat.hyperledger.org/channel/burrow-contributors
status: Draft
type: Non-Standards Track
category: Security, Keys management
created: 2019-08-21
---

## Securing validator and participants keys

This ADR introduces how we will ensure that private key matter is protected.

## Motivation

A burrow node can be a server for many clients; however, at the moment many clients use mempool signing
which requires the burrow node to hold the private keys for all participants, and all client needs to know
is the address of the account and access to the node's grpc port. This is not secure.

mempool signing means that signing must happen in-process in burrow. It would be much preferable if the burrow
process did not have access to the participants keys, and if the keys could be managed.

In addition, the keys server mixes validator keys with participants keys, so it is possible to sign a transaction
using the validator key.

A burrow network is blockchain; as such we expect authenticity but do not expect confidentiality.

Furthermore encryption is not a goal in it-self, however the private key matter must be protected, so there
should be a tls and basic auth available for the keys server.

## Specification

This change requires changes to many components. We list them in order of implementation.

## burrow keys service no longer stores validator key

The validator key should simply be stored in a file and tenderminet should not use the keys service. 

- This allows the key to be management by tendermint, using e.g. their kms service
- This separates the validator keys from the participants keys
- The validator key and node key should be managed by a higher authority than the rest of the keys

## burrow keys should maintain address sequence number

Clients like burrow-js should not have to aware of an accounts sequence number. This was previously done by
mempool signing, but this requires the mempoool to have access to all the private keys. So, we move signing
to the keys service.

The keys service will have SignTx(payload.Tx) endpoint rather than a plain Sign() endpoint. This means that
the keys service is aware of what it is signing and can populate the sequence number if it is provided with
a sequence number of 0.

As a result, the keys service has to be aware of what sequence numbers the accounts are at. This requires
some replicated state. burrow keys should retrieve the account sequence number from burrow if needed. This
is also the case if the data the keys service holds is older than 5 seconds.

*FIXME* Some transactions increase the sequence number by more than 1. Not sure what they were and if this
still is the case.

## burrow.js

burrow js should require a keys server and use SignTx() rather than mempool signing.

The keys server can be the same as the burrow node if using the built-in keys service.

## burrow deploy

Same changes as burrow js.

## burrow node

burrow should not longer permit mempool signing; no delegate signing. The keys service will provide this
functionality. 

In addition, burrow should allow sequence number which are monotonically increasing. The keys service does
not know which transactions succeed and which will fail, so it is possible that the sequence number maintained
in the key service will get out of sync with the chain.

In a dump-restore scenario the sequence numbers will also change.

## Encryption and authentication to the key service (TLS and basic auth)

The keys service should provide a mode where it is available in via an encrypted and authenicated setup.
This is not required for our own infrastructure.
