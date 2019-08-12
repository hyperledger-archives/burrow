/**
 * @file Burrow.js
 * @fileOverview Factory module for the Burrow class.
 * @author Dennis Mckinnon
 * @module Burrow
 */

'use strict'

var Service = require('./service')
var events = require('./events.js')

var Pipe = require('./pipe')
var ContractManager = require('./contractManager')
var Namereg = require('./namereg')

/**
 * Create a new instance of the Burrow class.
 *
 * @param {string} URL - URL of Burrow instance.
 * @returns {Burrow} - A new instance of the Burrow class.
 */
exports.createInstance = function (URL, account, options) {
  URL = (typeof URL === 'string' ? URL : URL.host + ':' + URL.port)
  return new Burrow(URL, account, options)
}

/**
 * The main class.
 *
 * @param {string} URL - URL of Burrow instance.
 * @constructor
 */
function Burrow (URL, account, options) {
  this.URL = URL
  this.tag = options.tag

  if (!account) {
    this.readonly = true
    this.account = null
  } else {
    this.readonly = false
    this.account = account
  }

  this.executionEvents = Service('rpcevents.proto', 'rpcevents', 'ExecutionEvents', URL)

  this.transact = Service('rpctransact.proto', 'rpctransact', 'Transact', URL)
  this.query = Service('rpcquery.proto', 'rpcquery', 'Query', URL)

  // This is the execution events streaming service running on top of the raw streaming function.
  this.events = events(this)

  // Contracts stuff running on top of grpc
  this.pipe = new Pipe(this)
  this.contracts = new ContractManager(this, options)

  this.namereg = new Namereg(this)
}
