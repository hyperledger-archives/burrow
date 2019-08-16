/**
 * @file contractManager.js
 * @author Dennis Mckinnon
 * @date 2018
 * @module contracts
 */

var Contract = require('./contracts/contract')

var defaultHandlers = {
  call: function (result) {
    // console.log(result)
    return result.raw
  },
  con: function (result) {
    // console.log(result)
    return result.contractAddress
  }
}

var nullHandler = function (result) { return result }

function ContractManager (burrow, options) {
  options = Object.assign({ objectReturn: false }, options)
  var handlers = Object.assign({}, defaultHandlers, options.handlers)
  this.burrow = burrow
  this.handlers = handlers

  // As of 0.25.0 change the default handler to use nullhandler by default
  if (options.objectReturn) {
    this.handlers.call = nullHandler
  } else {
    console.log('DEPRECATION WARNING. As of 0.25.0 the default behaviour of contract calls will be to return the full result object (instead of an array of arguments)')
    console.log('If you wish to keep this behaviour after 0.25.0 you can recreate it by using a handler function for calls.')
    console.log('This can be done by passing {handlers: {call: function(result){return result.raw}} as an option to burrow object creation (new burrow(URL, account, options))')
  }
}

/**
 * Should be called to create new contract on a blockchain
 *
 * @method new
 * @param {Object} abi object (required)
 * @param {string} byteCode - Hex encoded bytecode of contact
 * @param {*} [contract] constructor param1 (optional)
 * @param {*} [contract] constructor param2 (optional)
 * @param {Function} callback (optional)
 * @param {Object} Handlers (optional)
 * @returns {Contract} returns a promise if no callback provided
 */
ContractManager.prototype.deploy = function () {
  // parse arguments
  var callback = null
  var handlers = Object.assign({}, this.handlers)

  var args = Array.prototype.slice.call(arguments)
  if (args[args.length - 1] instanceof Object) {
    handlers = Object.assign(handlers, args.pop())
  }

  if (args[args.length - 1] instanceof Function) {
    callback = args.pop()
  }

  // TODO just pass in the bytecode and set it don't do this merging
  var abi = args.shift()
  var byteCode = args.shift()

  var contract = new Contract(abi, null, byteCode, this.burrow, handlers)
  var P = contract._constructor.apply(contract, args).then((address) => { contract.address = address; return contract })

  if (callback) {
    P.then((contract) => { return callback(null, contract) })
      .catch((err) => { return callback(err) })
  } else {
    return P.then(() => { return contract })
  }
}

/**
 * Creates a contract object interface from an abi
 *
 * @method new
 * @param {Object} abi - abi object for contract
 * @param {string} byteCode - Hex encoded bytecode of contact [can be null]
 * @param {string} address - default contract address [can be null]
 * @returns {Contract} returns contract interface object
 */
ContractManager.prototype.new = function (abi, byteCode, address, handlers) {
  handlers = Object.assign({}, this.handlers, handlers)
  return new Contract(abi, address, byteCode, this.burrow, handlers)
}

module.exports = ContractManager
