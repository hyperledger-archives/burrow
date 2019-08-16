'use strict'

const assert = require('assert')
const test = require('../../lib/test')

const Test = test.Test()

describe('#45', function () {
  before(Test.before())
  after(Test.after())

  this.timeout(10 * 1000)

  it('nottherealbatman', Test.it(function (burrow) {
    const source = `
      pragma solidity ^0.4.21;
      contract Test {
          string _name;

          function add(int a, int b) public pure returns (int sum) {
              sum = a + b;
          }

          function setName(string newname) public {
             _name = newname;
          }

          function getName() public view returns (string) {
              return _name;
          }
      }
    `
    const {abi, bytecode} = test.compile(source, ':Test')
    return burrow.contracts.deploy(abi, bytecode).then((contract) =>
      contract.setName('Batman')
        .then(() => contract.getName())
    ).then((value) => {
      assert.equal(value, 'Batman')
    })
  }))

  it('rguikers', Test.it(function (burrow) {
    const source = `
      pragma solidity ^0.4.21;
      contract Test {

          function getAddress() public view returns (address) {
            return this;
          }

          function getNumber() public pure returns (uint) {
            return 100;
          }

      }
    `

    const {abi, bytecode} = test.compile(source, ':Test')
    return burrow.contracts.deploy(abi, bytecode).then((contract) =>
      Promise.all([contract.getAddress(), contract.getNumber()])
        .then(([address, number]) => {
          assert.equal(address[0].length, 40)
          assert.equal(number[0], 100)
        })
    )
  }))
})
