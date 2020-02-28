import * as assert from 'assert';
import * as test from '../test';

const Test = test.Test();

describe('#47', function () {
  before(Test.before())
  after(Test.after())

  this.timeout(10 * 1000)

  it('#47', Test.it(function (burrow) {
    const source = `
      pragma solidity >=0.0.0;
      contract Test{
        string _withSpace = "  Pieter";
        string _withoutSpace = "Pieter";

        function getWithSpaceConstant() public view returns (string memory) {
          return _withSpace;
        }

        function getWithoutSpaceConstant () public view returns (string memory) {
          return _withoutSpace;
        }
      }
    `
    const {abi, bytecode} = test.compile(source, 'Test')
    return burrow.contracts.deploy(abi, bytecode)
      .then((contract: any) => Promise.all([contract.getWithSpaceConstant(), contract.getWithoutSpaceConstant()]))
      .then(([withSpace, withoutSpace]) => {
        assert.equal(withSpace, '  Pieter')
        assert.equal(withoutSpace, 'Pieter')
      })
  }))
})
