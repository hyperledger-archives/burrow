// Package consensus contains the logic maintaining and connecting to our Tendermint
// consensus network. The job of the consensus module is to establish a total order
// of transactions shared by all validators, which can then be executed over by the
// execution module.
package consensus
