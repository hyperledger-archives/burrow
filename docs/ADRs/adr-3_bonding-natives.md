---
adr: 3
title: Smart Contract Bonding Natives
author: Silas Davis (@silasdavis), Gregory Hill (@gregdhill)
discussions-to: https://chat.hyperledger.org/channel/burrow-contributors
status: Draft
type: Standards Track
category: State, Consensus, Governance
created: 2019-07-08
---

## Abstract

In vanilla proof-of-stake, an account with some amount of token pledges to vest a portion of this to actively participate in consensus with the knowledge that
misbehavior could result in punishment - bonded power is slashed to some extent. [PR 1100](https://github.com/hyperledger/burrow/pull/1100) contains the base
implementation to support this model, but we foresee techniques such as delegation being important to network users in the future. Therefore, we propose a 
smart contract orientated approach which leverages SNatives to expose 'admin' functionality for controlling individual validator investments.

## Motivation

There are countless ways to model token economics, even in Proof-of-Stake (PoS) there are a number of schemes such as delegation, nomination or even hybrid
approaches. It is conceivable that we may want to incorporate alternate methods in the future without forking which (depending on the technique) may not be
easily done. Outsourcing this task to individual validators makes sense in the same way we do not control how native accounts move and use their tokens.

## Specification

A management contract should sit at the address of each validator bonded onto the network which contains the logic for how that validator may operate. For instance,
delegation would be trivial to implement if we could simply transfer funds to this account and have the smart contract automatically bond them. This special account
then handles the validators portfolio at its discretion, simplifying our consensus overhead with a tight account to validator binding. 

1. Account w/ bond permission signs and sends BondTx
    - (Optionally) add EVM / WASM bytecode
2. Bond given amount for account into validator set
3. Call against validator address checks for existence of code and executes 
4. Exposed natives verify validator address / power and run directly against state, updating or removing power

We would still like to maintain the notion of validator 'flow' to ensure that the set does not change too quickly. Additionally, if power is ever depleted then we
may want to consider whether we should retain the contract. This problem also lends itself to upgradeability as a poorly written contract could severely impact the
lifetime and reputation of the identity in question. One possible solution is to hard-code the special validator contract and make it a concern of the network - 
upgradeable through governance / proposals. This has the benefit of equalizing all validators on a per-chain basis and makes patching vulnerabilities easier.