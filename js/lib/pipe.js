/**
 * @file dev_pipe.js
 * @fileOverview Base class for the dev-pipe.
 * @author Andreas Olofsson
 * @module pipe/dev_pipe
 */
'use strict'

/**
 * Constructor for the Pipe class.
 *
 * @type {Pipe}
 */
module.exports = Pipe

/**
 * DevPipe transacts using the unsafe private-key transactions.
 *
 * @param {*} burrow - the burrow object.
 * @param {string} accounts - the private key to use when sending transactions. NOTE: This means a private key
 * will be passed over the net, so it should only be used when developing, or if it's 100% certain that the
 * Burrow server and this script runs on the same machine, or communication is secure. The recommended way
 * will be to call a signing function on the client side, like in a browser plugin.
 *
 * @constructor
 */
function Pipe (burrow) {
  this.burrow = burrow
}

/**
 * Used to send a transaction.
 * @param {module:solidity/function~TxPayload} txPayload - The payload object.
 * @param callback - The error-first callback. The 'data' param is a contract address in the case of a
 * create transactions, otherwise it's the return value.
 */
Pipe.prototype.transact = function (txPayload, callback) {
  this.burrow.transact.CallTxSync(txPayload, callback)
}

/**
 * Used to do a call.
 * @param {module:solidity/function~TxPayload} txPayload - The payload object.
 * @param callback - The error-first callback.
 */
Pipe.prototype.call = function (txPayload, callback) {
  this.burrow.transact.CallTxSim(txPayload, callback)
}

/**
 * Used to subscribe to Solidity events from a given account.
 *
 * @param {string} accountAddress - the address of the account.
 * @param {function} createCallback - error-first callback. The data object is the EventSub object.
 * @param {function} eventCallback - error-first callback. The data object is a solidity event object.
 */
Pipe.prototype.eventSub = function (accountAddress, signature, callback) {
  return this.burrow.events.subContractEvents(accountAddress, signature, {}, callback)
}

Pipe.prototype.burrow = function () {
  return this.burrow
}
