'use strict'

const assert = require('assert')
const test = require('../../lib/test')

const Test = test.Test()

describe('#46', function () {
  before(Test.before())
  after(Test.after())

  this.timeout(10 * 1000)

  it('#46', Test.it(function (burrow) {
    const source = `
      pragma solidity ^0.4.21;
      contract Test{

        string _name;

        function setName(string newname) public {
          _name = newname;
        }

        function getNameConstant() public view returns (string) {
          return _name;
        }

        function getName() public view returns (string) {
          return _name;
        }
      }
    `

    const {abi, bytecode} = test.compile(source, ':Test')
    return burrow.contracts.deploy(abi, bytecode)
      .then((contract) => contract.setName('Batman')
        .then(() => Promise.all([contract.getNameConstant(), contract.getName()])))
      .then(([constant, nonConstant]) => {
        assert.equal(constant[0], nonConstant[0])
      })
  }))
})
