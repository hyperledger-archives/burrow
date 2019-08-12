/**
 * @file sha3.js
 * @author Marek Kotewicz <marek@ethdev.com>
 * @author Andreas Olofsson
 * @date 2015
 * @module utils/sha3
 */
var utils = require('./utils')
var sha3 = require('crypto-js/sha3')

module.exports = function (str, isNew) {
  if (str.substr(0, 2) === '0x' && !isNew) {
    console.warn('requirement of using web3.fromAscii before sha3 is deprecated')
    console.warn('new usage: \'web3.sha3("hello")\'')
    console.warn('see https://github.com/ethereum/web3.js/pull/205')
    console.warn('if you need to hash hex value, you can do \'sha3("0xfff", true)\'')
    str = utils.toAscii(str)
  }

  return sha3(str, {
    outputLength: 256
  }).toString().toUpperCase()
}
