import * as assert from 'assert';
import * as test from '../test';

const Test = test.Test();

describe('#48', function () {
  before(Test.before())
  after(Test.after())

  this.timeout(10 * 1000)

  it('#48', Test.it(function (burrow) {
    const source = `
      pragma solidity >=0.0.0;
      contract Test {

          function getAddress() public view returns (address) {
            return address(this);
          }

          function getNumber() public pure returns (uint) {
            return 100;
          }

          function getCombination() public view returns (uint _number, address _address) {
            _number = 100;
            _address = address(this);
          }

      }
    `
    const {abi, bytecode} = test.compile(source, 'Test')
    return burrow.contracts.deploy(abi, bytecode)
      .then((contract: any) => contract.getCombination())
      .then(([number, address]) => {
        assert.equal(number, 100)
        assert.equal(address.length, 40)
      })
  }))
})
