# @hyperledger/burrow

This is a TypeScript API for communicating with a [Hyperledger Burrow](https://github.com/hyperledger/burrow) server, which implements the GRPC spec.

[![npm version][npm-image]][npm-url]

## Version compatibility

This lib's version is pegged to burrow's version on the minor. So @hyperledger/burrow at version X.Y.^ will work with burrow version X.Y.^ where ^ means latest patch version. The patch version numbering will not always correspond. If you are having difficulties getting this lib to work with a burrow release please first make sure you have the latest patch version of each.