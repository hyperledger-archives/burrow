// +build forensics

// This package contains tools for examining, replaying, and debugging Tendermint-side and Burrow-side blockchain state.
// Some code is quick and dirty from particular investigations and some is better extracted, encapsulated and generalised.
// The sketchy code is included so that useful tools can be progressively put together as the generality of the types of
// forensic debugging needed in the wild are determined.

package forensics
