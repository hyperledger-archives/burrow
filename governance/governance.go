// The governance package contains functionality for altering permissions, token distribution, consensus parameters,
// validators, and network forks.
package governance

// TODO:
// - Set validator power
// - Set account amount(s)
// - Set account permissions
// - Set global permissions
// - Set ConsensusParams
// Future considerations:
// - Handle network forks/termination/merging/replacement ?
// - Provide transaction in stasis/sudo (voting?)
// - Handle bonding by other means (e.g. pre-shared key permitting n bondings)
// - Network administered proxies (i.e. instead of keys have password authentication for identities - allow calls to originate as if from address without key?)
// Subject to:
// - Less than 1/3 validator power change per block
