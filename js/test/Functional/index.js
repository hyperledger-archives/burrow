'use strict'

const assert = require('assert')
const test = require('../../lib/test')

const Test = test.Test()

describe('Functional Contract Usage', function () {
  before(Test.before())
  after(Test.after())

  this.timeout(10 * 1000)

  it('#Constructor usage', Test.it(function (burrow) {
    const source = `
      pragma solidity ^0.4.21;
      contract Test {
        address storedData;
        constructor(address x) public {
          storedData = x;
        }

        function getAddress() public view returns (address) {
          return this;
        }

        function getNumber() public pure returns (uint) {
          return 100;
        }

        function getCombination() public view returns (uint _number, address _address, string _saying, bytes32 _randomBytes, address _stored) {
          _number = 100;
          _address = this;
          _saying = "hello moto";
          _randomBytes = 0xDEADBEEFFEEDFACE;
          _stored = storedData;
        }

      }
    `
    const {abi, bytecode} = test.compile(source, ':Test')
    const contract = burrow.contracts.new(abi, bytecode)

    let A1
    let A2

    assert.equal(contract.address, null)

    // Use the _contructor method to creat two contracts
    return Promise.all(
      [contract._constructor('88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F'),
        contract._constructor('ABCDEFABCDEFABCDEFABCDEFABCDEFABCDEF0123')])
      .then(([address1, address2]) => {
        assert.equal(contract.address, null)
        A1 = address1
        A2 = address2

        return Promise.all(
          [contract.getCombination.at(A1),
            contract.getCombination.at(A2)])
      })
      .then(([object1, object2]) => {
        const expected1 = [100, A1, 'hello moto', '000000000000000000000000000000000000000000000000DEADBEEFFEEDFACE', '88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F']
        const expected2 = [100, A2, 'hello moto', '000000000000000000000000000000000000000000000000DEADBEEFFEEDFACE', 'ABCDEFABCDEFABCDEFABCDEFABCDEFABCDEF0123']
        assert.deepEqual(object1, expected1)
        assert.deepEqual(object2, expected2)
      })
  }))
})
