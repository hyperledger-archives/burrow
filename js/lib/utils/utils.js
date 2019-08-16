/**
 * @file utils.js
 * @author Marek Kotewicz <marek@ethdev.com>
 * @author Andreas Olofsson
 * @date 2015
 * @module utils/utils
 */

/**
 * Should be used to create full function/event name from json abi
 *
 * @method transformToFullName
 * @param {Object} json - json-abi
 * @return {String} full fnction/event name
 */
var transformToFullName = function (json) {
  if (json.name.indexOf('(') !== -1) {
    return json.name
  }

  var typeName = json.inputs.map(function (i) {
    return i.type
  }).join()
  return json.name + '(' + typeName + ')'
}

/**
 * Should be called to get display name of contract function
 *
 * @method extractDisplayName
 * @param {String} name of function/event
 * @returns {String} display name for function/event eg. multiply(uint256) -> multiply
 */
var extractDisplayName = function (name) {
  var length = name.indexOf('(')
  return length !== -1 ? name.substr(0, length) : name
}

/**
 *
 * @param {String} name - the name.
 * @returns {String} overloaded part of function/event name
 */
var extractTypeName = function (name) {
  /// TODO: make it invulnerable
  var length = name.indexOf('(')
  return length !== -1 ? name.substr(length + 1, name.length - 1 - (length + 1)).replace(' ', '') : ''
}

/**
 * Returns true if object is function, otherwise false
 *
 * @method isFunction
 * @param {Object} object - object to test
 * @return {Boolean}
 */
var isFunction = function (object) {
  return typeof object === 'function'
}

module.exports = {
  transformToFullName: transformToFullName,
  extractDisplayName: extractDisplayName,
  extractTypeName: extractTypeName,
  isFunction: isFunction
}
