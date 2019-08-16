# @monax/burrow (Alpha)

This is a JavaScript API for communicating with a [Hyperledger Burrow](https://github.com/hyperledger/burrow) server, which implements the GRPC spec.

[![npm version][npm-image]][npm-url]

## New Library

Previously our client libs were broken into two components `@monax/legacy-db.js` and `@monax/legacy-contract.js`. These have both been replaced by this library `@monax/burrow`. This upgrade was part of a major re-write on the back-end and as such ONLY `@monax/burrow` SHOULD BE USED WITH BURROW VERSIONS GREATER THAN 0.20.0. There is NO BACKWARDS COMPATIBILITY of this lib with versions of burrow less than 0.20.0. There will be a short guide below for upgrading existing applications to new burrow versions.

## Version compatibility

This lib's version is pegged to burrow's version on the minor. So @monax/burrow at version X.Y.^ will work with burrow version X.Y.^ where ^ means latest patch version. The patch version numbering will not always correspond. If you are having difficulties getting this lib to work with a burrow release please first make sure you have the latest patch version of each.