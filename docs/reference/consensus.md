# Consensus

Burrow uses the [Tendermint](https://tendermint.com/) consensus engine to provide a total ordering of all execution events. 
Tendermint is a partially synchronous Byzantine Fault Tolerant (BFT) consensus algorithm in the PBFT family. It provides us with:

- Total ordering of input transactions via the consensus algorithm itself
- Raw block storage and blockchain
- P2P layer - gossip of blocks, transactions, and votes

## Quorum

Tendermint requires strictly greater than 2/3 of its validator voting power to be online at all times to make progress and requires that strictly less than 
1/3 of its validators are behaving arbitrarily - e.g. maliciously to maintain its safety and liveness guarantees.

Restated, we need 2F + 1 validators in order to tolerate F byzantine faults. Which is the tight optimal lower bound on the number of validators for consensus 
in the partially synchronous byzantine-fault model.

In the context of Tendermint where we have a potentially uneven distribution of voting power amongst validators this bound still applies except we treat each 
unit of voting power (which is an integral number) as 'a validator' and so provided strictly less than 1/3 of total power is in the hands of misbehaving validators 
consensus will make progress and remain consistent (i.e. all correct validators will store the identical correct state).

## How we use Tendermint

Burrow and Tendermint have their origins in a single codebase - 'eris-db' - where they were developed in tandem. Tendermint was spun off into a separate project and company 
as a modular reusable component focused exclusively on providing a consensus network layer. It consumes a network configuration and topology (described via its own genesis schema), 
provides the network connectivity, and consensus algorithm state machine for performing deterministic replication of state across all nodes connected to the network.

It does this replication by providing a stream of identically ordered transactions to each node via an protocol called the Application BlockChain Interface (ABCI). 
As far as Tendermint is concerned these transactions are opaque binary blobs, it is down to the 'application' to decide how to deserialise, verify, and apply the 
transactions to state. From Tendermint's perspective Burrow acts as an ABCI client.

Tendermint is designed so it can be run as a standalone service much like a database, in this mode of operation clients send transactions to Tendermint to be included in blocks 
over its RPC layer and ABCI clients receive requests over the ABCI wire protocol to validate transactions before and after they are included in blocks, with the ability to reject them.
Burrow does not use Tendermint in this way but instead uses Tendermint as an embedded library. Internally we still use the underlying interfaces of `BroadcastTx` and the ABCI interface 
but we use them within the context of single Go process. This situation owes something to our shared history, but also provides a number of benefits:

- We can be more efficient by communicating in-process rather than over RPC layers
- Users of Burrow do not need to run multiple services in order to form a network
- In being more tightly coupled to Tendermint we gain direct programmatic access to low-level aspects of its functioning like the mempool and P2P layers
- By controlling our own internal Tendermint node we can provide a consistent configuration for Burrow and Tendermint based on certain conventions and reflection
- We can provide control over validators, genesis, and network formation that would be harder to do with separate services
- We can provide a single consistent command line interface

The upshot of all of this is that while Tendermint is a crucial part of the Burrow implementation and Burrow is intended to keep and build on its compatibility with public Tendermint
and Cosmos networks it is essentially an implementation detail of Burrow. Burrow is not intended to function as a standalone ABCI app.


### What we do not use from Tendermint

There are a number of pieces of functionality you can find documented as part of 'Tendermint the blockchain' that we opt not use as part of 'Tendermint the library'. 
These points of difference are not just for difference's sake but but because we can provide a more specific coherent version of them as part of Burrow which has its 
own needs in each of the following areas:

- The Tendermint HTTP RPC layers - all of our write RPC layers are GRPC-based
- The Tendermint transaction index - we have a GRPC execution events service tailored to our own [state](/docs/reference/state.md)
- Tendermint configuration - we re-expose certain configuration options from Tendermint combined with our own - some of which control multiple Tendermint parameters
- Tendermint genesis - we provide our own genesis from which we derive Tendermint's genesis doc


### The Burrow/Tendermint divide

Tendermint maintains its own on-disk state consisting of:

- Consensus state
- Write-Ahead log (WAL)
- Validator set
- Block store

It is able to use these to recover from (in principle) crash-failures, whereby it can detect faulty data (within certain physical bounds) and recover by connecting to 
other nodes on the network to 'catch up' by replaying blocks. When a node first starts it will enter 'fast sync' mode and catch up to the latest block height on the network.

We consider the consensus state entirely the domain of Tendermint - we rely upon Tendermint to order transactions for us. We rely upon this and the WAL to give Burrow 
crash-fault tolerance - Burrow maintains a checkpoint that runs one block behind, if Burrow was to crash, Tendermint should recover and reply the block overwriting any 
potentially dirty writes for the block that was in-flight.

The validator set is something that is administered by Burrow and for which Tendermint necessarily keeps its own accounting. We therefore have representations for the current 
validator state stored both in Burrow's core code and in our Tendermint library code - which we ensure are synchronised. We furthermore store a complete history of previous validator sets.

We have a similar arrangement with the block store - Tendermint is responsible for storing the raw block data which includes the full array of votes and headers relating to consensus 
and the raw serialised transaction data. However Burrow also has much richer execution event data that comprises a complete trace of all execution events stored as a stream in a merkle tree. 
This means that there is some redundancy of data between the Tendermint side and Burrow side storage of transactions and blocks. One way to think about this is as Tendermint providing the 
low level transaction log (as a relational database might) and Burrow the application level schema and indices. We also support a mode of operation (currently just for a single node) 
involving no consensus (by setting `Enabled = false` in the Tendermint section of the Burrow configuration) where our state storage is exclusive. There are future possibilities enabled 
by being able to operate without Tendermint including for private state channels and alternative consensus mechanisms.

For more details see our [state documentation](/reference/state.md).
