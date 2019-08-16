'use strict'

const assert = require('assert')
const test = require('../../lib/test')

const Test = test.Test()

describe('REVERT constant', function () {
  this.timeout(10 * 1000)
  let contract

  before(Test.before(function (burrow) {
    const source = `
      pragma solidity ^0.4.21;
      contract c {
        string s = "secret";
        function getString(uint key) public constant returns (string){
          if (key != 42){
            revert("Did not pass correct key");
          } else {
            return s;
          }
        }
      }
    `

    const {abi, bytecode} = test.compile(source, ':c')
    return burrow.contracts.deploy(abi, bytecode).then((c) => {
      contract = c
    })
  }))

  after(Test.after())

  it('gets the string when revert not called', Test.it(function () {
    return contract.getString(42)
      .then((str) => {
        assert.equal(str, 'secret')
      })
  }))

  it('It catches a revert with the revert string',
    Test.it(function () {
      return contract.getString(1)
        .then((str) => {
          throw new Error('Did not catch revert error')
        }).catch((err) => {
          assert.equal(err.code, 'ERR_EXECUTION_REVERT')
          assert.equal(err.message, 'Did not pass correct key')
        })
    })
  )
})
