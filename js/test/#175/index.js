'use strict'

const test = require('../../lib/test')
const assert = require('assert')

const Test = test.Test()

describe('#175', function () {
  before(Test.before())
  after(Test.after())

  this.timeout(10 * 1000)

  it('#175', Test.it(function (burrow) {
    const source = `
      pragma solidity ^0.4.21;
      contract Contract {
        string thename;
        constructor(string newName) public {
          thename = newName;
        }
        function getName() public view returns (string name) {
          return thename;
        }
      }
    `
    var contract
    let A2

    const {abi, bytecode} = test.compile(source, ':Contract')
    return burrow.contracts.deploy(abi, bytecode, 'contract1').then((C) => {
      contract = C
      return contract._constructor('contract2')
    }).then((address) => {
      A2 = address
      return Promise.all(
        [contract.getName(),      // Note using the default address from the deploy
          contract.getName.at(A2)])   // Using the .at() to specify the second deployed contract
    }).then(([result1, result2]) => {
      assert.equal(result1[0], 'contract1')
      assert.equal(result2[0], 'contract2')
    })
  }))
})
