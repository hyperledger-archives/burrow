# Bond / Unbond

Burrow adopts the notion of Bonded Proof-of-Stake (BPoS). This means that in order to 
participate in consensus, a node must give some value of its token as collateral. 
The more token that is bonded, the higher the chance of 'mining' a block.

## How it Works

When starting a burrow node, we provide an address which links to an owned key-pair.
Each running node thus has an identity which may active in the validator set. Assuming we have
connected to a network, our node will replay state downloaded from its peers - from which we will
be able to discern any native token stored at our address. If this amount is non-negligible
we can submit a signed BondTx to be gossiped amongst the current validators who should include it
in a block. When executing this transaction against our global state, burrow will first check that the 
input address has correctly signed the payload, that the respective account exists with the bonding
permission and has enough token to stake. If successful, then this will subtract the token from the
account and raise the new validators power - enabling it to vote and propose new blocks. The procedure 
for unbonding is antithetical, diminishing the validator accounts power on success.

One nuance with altering the validator set is to do with a concept we call the 'max flow'.
To prevent the validator pool changing too quickly over a single block whilst ensuring the 
majority of validators are non-byzantine after the transition, we allow up to `ceil((t)/3) - 1`
to be changed where `t` is the current total validator power.


## Future Work

Currently a validator must bond or unbond themselves directly - we enforce a strict relationship 
that prohibits staking token to another account. Funds may be send to another account however, 
and providing that also has permission to bond, then it may join the validator set. In the future we
hope to extend our model to allow for delegation, whereby any party with native token may stake it
against a running validator and receive a share of the rewards.