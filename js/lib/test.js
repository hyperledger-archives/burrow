'use strict'

const Burrow = require('..')
const url = require('url')
const Solidity = require('solc')

const blockchainUrl = (urlObj) => {
  var envUrl = {}
  if (process.env.BURROW_GRPC_PORT) envUrl.port = process.env.BURROW_GRPC_PORT
  if (process.env.BURROW_HOST) envUrl.hostname = process.env.BURROW_HOST

  urlObj = Object.assign({port: '20997', hostname: '127.0.0.1'}, envUrl, urlObj)
  return url.format(urlObj)
}

// Convenience function to compile Solidity code in tests.
const compile = (source, name) => {
  const compiled = Solidity.compile(source, 1)
  if (compiled.errors) {
    throw new Error(compiled.errors)
  }
  const contract = compiled.contracts[name]
  const abi = JSON.parse(contract.interface)
  const bytecode = contract.bytecode

  return {abi, bytecode}
}

// Return a contract manager in the test harness.
const Test = (options) => {
  options = options || {}
  const urlString = (typeof options.url === 'string' ? options.url : blockchainUrl(options.url))
  delete options.url
  let account
  let burrow

  return {
    before: (burrowOptions, callback) =>
      function () {
        if (typeof burrowOptions === 'function') {
          callback = burrowOptions
          burrowOptions = {}
        }
        if (!burrowOptions) burrowOptions = {}

        try {
          account = JSON.parse(process.env.account)
        } catch (err) {
          return Promise.reject(new Error('Could not parse required account JSON: ' + process.env.account + ' Make sure you are passing a valid account json string as an env var account=\'{accountdata}\''))
        }

        // Options overrules defaults
        // burrowOptions = Object.assign(burrowOptions, options)

        burrow = Burrow.createInstance(urlString, account.address, burrowOptions)

        if (callback) {
          return callback(burrow) // eslint-disable-line
        }
      },

    it: (callback) =>
      function () {
        return callback(burrow) // eslint-disable-line
      },

    after: () =>
      function () {
      }
  }
}

module.exports = {
  compile,
  Test
}
