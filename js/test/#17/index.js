'use strict'

const assert = require('assert')
const test = require('../../lib/test')

const Test = test.Test()

describe('#17', function () {
  before(Test.before({handlers: {call: function (result) { return {super: result.values, man: result.raw} }}}))
  after(Test.after())

  this.timeout(10 * 1000)

  it('#17 Testing Per-contract handler overwriting', Test.it(function (burrow) {
    const source = `
      pragma solidity ^0.4.21;
      contract Test {

          function getAddress() public view returns (address) {
            return this;
          }

          function getNumber() public pure returns (uint) {
            return 100;
          }

          function getCombination() public view returns (uint _number, address _address, string _saying, bytes32 _randomBytes) {
            _number = 100;
            _address = this;
            _saying = "hello moto";
            _randomBytes = 0xDEADBEEFFEEDFACE;
          }

      }
    `
    let address
    const {abi, bytecode} = test.compile(source, ':Test')
    return burrow.contracts.deploy(abi, bytecode, {call: function (result) { return {values: result.values, raw: result.raw} }}).then((contract) => {
      address = contract.address
      return contract.getCombination()
    }).then((returnObject) => {
      const expected = {
        values: {
          _number: 100,
          _address: address,
          _saying: 'hello moto',
          _randomBytes: '000000000000000000000000000000000000000000000000DEADBEEFFEEDFACE'
        },
        raw: [100, address, 'hello moto', '000000000000000000000000000000000000000000000000DEADBEEFFEEDFACE']
      }
      assert.deepEqual(returnObject, expected)
    })
  }))
})
