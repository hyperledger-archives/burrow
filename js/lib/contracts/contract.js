'use strict'

var SolidityEvent = require('./event')
var SolidityFunction = require('./function')

/**
 * The contract type. This class is instantiated internally through the factory.
 *
 * @method Contract
 * @param {Array} abi
 * @param {string} address
 * @param {pipe} pipe;
 * @param {Function} outputFormatter - the output formatter.
 */
var Contract = function (abi, address, code, burrow, handlers) {
  this.address = address
  this.abi = abi
  this.code = code
  this.burrow = burrow
  this.handlers = handlers

  addFunctionsToContract(this)
  addEventsToContract(this)
}

/**
 * Should be called to add functions to contract object
 *
 * @method addFunctionsToContract
 * @param {Contract} contract
 * @param {Array} abi
 * TODO
 * @param {function} pipe - The pipe (added internally).
 * @param {function} outputFormatter - the output formatter (added internally).
 */

var addFunctionsToContract = function (contract) {
  contract.abi.filter(function (json) {
    return (json.type === 'function' || json.type === 'constructor')
  }).forEach(function (json) {
    let {displayName, typeName, call, encode, decode} = SolidityFunction(json)

    if (json.type === 'constructor') {
      contract._constructor = call.bind(contract, false, contract.handlers.con, '')
    } else {
      // bind the function call to the contract, specify if call or transact is desired
      var execute = call.bind(contract, json.constant, contract.handlers.call, null)
      execute.sim = call.bind(contract, true, contract.handlers.call, null)
      // These allow the interface to be used for a generic contract of this type
      execute.at = call.bind(contract, json.constant, contract.handlers.call)
      execute.atSim = call.bind(contract, true, contract.handlers.call)

      execute.encode = encode.bind(contract)
      execute.decode = decode.bind(contract)

      // Attach to the contract object
      if (!contract[displayName]) {
        contract[displayName] = execute
      }
      contract[displayName][typeName] = execute
    }
  })

  // Not every abi has a constructor specification.
  // If it doesn't we force a _constructor with null abi
  if (!contract._constructor) {
    let {call} = SolidityFunction(null)
    contract._constructor = call.bind(contract, false, contract.handlers.con, '')
  }
}

/**
 * Should be called to add events to contract object
 *
 * @method addEventsToContract
 * @param {Contract} contract
 * @param {Array} abi
 */
var addEventsToContract = function (contract) {
  contract.abi.filter(function (json) {
    return json.type === 'event'
  }).forEach(function (json) {
    let {displayName, typeName, call} = SolidityEvent(json)

    var execute = call.bind(contract, null)
    execute.once = call.bind(contract, null)
    execute.at = call.bind(contract)
    if (!contract[displayName]) {
      contract[displayName] = execute
    }
    contract[displayName][typeName] = call.bind(contract)
  })
}

module.exports = Contract
