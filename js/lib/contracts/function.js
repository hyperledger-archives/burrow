var utils = require('../utils/utils')
var convert = require('../utils/convert')
// var formatters = require('./formatters');
var sha3 = require('../utils/sha3')
var coder = require('ethereumjs-abi')

var config = require('../utils/config')
var ZERO_ADDRESS = Buffer.from('0000000000000000000000000000000000000000', 'hex')

var functionSig = function (abi) {
  var name = utils.transformToFullName(abi)
  return sha3(name).slice(0, 8)
}

var types = function (args) {
  return args.map(function (arg) {
    return arg.type
  })
}

var txPayload = function (data, account, address) {
  var payload = {}

  payload.Input = {Address: Buffer.from(account || ZERO_ADDRESS, 'hex'), Amount: 0}
  payload.Address = address ? Buffer.from(address, 'hex') : null
  payload.GasLimit = config.DEFAULT_GAS
  payload.Fee = 0
  payload.Data = Buffer.from(data, 'hex')

  return payload
}

var encodeF = function (abi, args, bytecode) {
  if (abi) {
    var abiInputs = types(abi.inputs)
    args = convert.burrowToAbi(abiInputs, args) // If abi is passed, convert values accordingly
  }

  // If bytecode provided then this is a creation call, bytecode goes first
  if (bytecode) {
    var data = bytecode
    if (abi) data += convert.bytesTB(coder.rawEncode(abiInputs, args))
    return data
  } else {
    return functionSig(abi) + convert.bytesTB(coder.rawEncode(abiInputs, args))
  }
}

var decodeF = function (abi, output) {
  if (!output) {
    return
  }

  var outputs = abi.outputs
  var outputTypes = types(outputs)

  // Decode raw bytes to arguments
  var raw = convert.abiToBurrow(outputTypes, coder.rawDecode(outputTypes, Buffer.from(output, 'hex')))

  // If an object is wanted,
  var result = {raw: raw.slice()}

  var args = outputs.reduce(function (acc, current) {
    var value = raw.shift()
    if (current.name) {
      acc[current.name] = value
    }
    return acc
  }, {})

  result.values = args

  return result
}

/**
 * Calls a contract function.
 *
 * @method call
 * @param {...Object} Contract function arguments
 * @param {function}
 * @return {String} output bytes
 */
var SolidityFunction = function (abi) {
  var isCon = (abi == null || abi.type === 'constructor')
  var name
  var displayName
  var typeName

  if (!isCon) {
    name = utils.transformToFullName(abi)
    displayName = utils.extractDisplayName(name)
    typeName = utils.extractTypeName(name)
  }

  // It might seem weird to include copies of the functions in here and above
  // My reason is the code above can be used "functionally" whereas this version
  // Uses implicit attributes of this object.
  // I want to keep them separate in the case that we want to move all the functional
  // components together and maybe even... write tests for them (gasp!)
  var encode = function () {
    var args = Array.prototype.slice.call(arguments)
    return encodeF(abi, args, isCon ? this.code : null)
  }

  var decode = function (data) {
    return decodeF(abi, data, this.objectReturn)
  }

  var call = function () {
    var args = Array.prototype.slice.call(arguments)
    var isSim = args.shift()
    var handler = args.shift() || function (result) { return result }
    var address = args.shift() || this.address

    if (isCon) { address = null }

    var callback
    if (utils.isFunction(args[args.length - 1])) { callback = args.pop() };

    var self = this

    var P = new Promise(function (resolve, reject) {
      if (address == null && !isCon) reject(new Error('Address not provided to call'))
      if (abi != null && abi.inputs.length !== args.length) reject(new Error('Insufficient arguments passed to function: ' + (isCon ? 'constructor' : name)))
      // Post process the return
      var post = function (error, result) {
        if (error) return reject(error)

        // Handle execution reversions
        if (result.Exception && result.Exception.Code === 17) {
          // Decode error message if there is one otherwise default
          if (result.Result.Return.length === 0) {
            error = new Error('Execution Reverted')
          } else {
            // Strip first 4 bytes(function signature) the decode as a string
            error = new Error(coder.rawDecode(['string'], result.Result.Return.slice(4))[0])
          }
          error.code = 'ERR_EXECUTION_REVERT'
          return reject(error)
        }

        // Unpack return arguments
        var returnObj = {}
        if (!isCon) {
          try {
            returnObj = decodeF(abi, result.Result.Return, self.objectReturn)
          } catch (e) {
            return reject(e)
          }
        }

        // Meta Data (address, caller, height, etc)
        returnObj.contractAddress = result.Receipt.ContractAddress.toString('hex').toUpperCase()
        returnObj.height = result.Header.Height
        returnObj.index = result.Header.Index
        returnObj.hash = result.Header.TxHash.toString('hex').toUpperCase()
        returnObj.type = result.Header.TxType
        returnObj.result = result.Result
        returnObj.tx = result.Envelope
        returnObj.caller = convert.recApply(result.Envelope.Signatories, convert.bytesTB)

        return resolve(returnObj)
      }

      // Decide if to make a "call" or a "transaction"
      // For the moment we need to use the burrowtoweb3 function to prefix bytes with 0x
      // otherwise the coder will give an error with bugnumber not a number
      // TODO investigate if other libs or an updated lib will fix this
      // var data = encodeF(abi, utils.burrowToWeb3(args), isCon ? self.code : null)
      var data = encodeF(abi, args, isCon ? self.code : null)
      var payload = txPayload(data, self.burrow.account || ZERO_ADDRESS, address)

      if (isSim) {
        self.burrow.pipe.call(payload, post)
      } else {
        self.burrow.pipe.transact(payload, post)
      }
    })

    if (callback) {
      P.then(handler).then((result) => { return callback(null, result) })
        .catch((err) => { return callback(err) })
    } else {
      return P.then(handler)
    }
  }

  return {displayName, typeName, call, encode, decode}
}

module.exports = SolidityFunction
