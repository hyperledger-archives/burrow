'use strict'

const assert = require('assert')
const test = require('../../lib/test')

const Test = test.Test()

describe('#48', function () {
  before(Test.before())
  after(Test.after())

  this.timeout(10 * 1000)

  it('#48', Test.it(function (burrow) {
    const source = `
      pragma solidity ^0.4.21;
      contract Test {

          function getAddress() public view returns (address) {
            return this;
          }

          function getNumber() public pure returns (uint) {
            return 100;
          }

          function getCombination() public view returns (uint _number, address _address) {
            _number = 100;
            _address = this;
          }

      }
    `
    const {abi, bytecode} = test.compile(source, ':Test')
    return burrow.contracts.deploy(abi, bytecode)
      .then((contract) => contract.getCombination())
      .then(([number, address]) => {
        assert.equal(number, 100)
        assert.equal(address.length, 40)
      })
  }))
})
