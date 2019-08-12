'use strict'

const assert = require('assert')
const test = require('../../lib/test')

const Test = test.Test()

describe('HTTP', function () {
  before(Test.before())
  after(Test.after())

  this.timeout(10 * 1000)

  it('sets and gets a value from a contract', Test.it(function (burrow) {
    const source = `
      pragma solidity ^0.4.21;
      contract SimpleStorage {
          uint storedData;

          function set(uint x) public {
              storedData = x;
          }

          function get() constant public returns (uint retVal) {
              return storedData;
          }
      }
    `
    const {abi, bytecode} = test.compile(source, ':SimpleStorage')
    return burrow.contracts.deploy(abi, bytecode)
      .then((contract) => contract.set(42)
        .then(() => contract.get())
      ).then((value) => {
        assert.equal(value, 42)
      })
  }))
})
