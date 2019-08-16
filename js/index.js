/**
 * @file index.js
 * @fileOverview Index file for the Burrow javascript API. This file contains a factory method
 * for creating a new <tt>Burrow</tt> instance.
 * @author Andreas Olofsson
 * @module index
 */
'use strict'

var Burrow = require('./lib/Burrow')
var utils = require('./lib/utils/utils')

module.exports = {
  createInstance: Burrow.createInstance,
  Burrow: Burrow,
  utils: utils
}
