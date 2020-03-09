import * as assert from 'assert';
import * as test from '../test';

const Test = test.Test();

describe('HTTP', function () {
  before(Test.before())
  after(Test.after())

  this.timeout(10 * 1000)

  it('sets and gets a value from a contract', Test.it(function (burrow) {
    const source = `
      pragma solidity >=0.0.0;
      contract SimpleStorage {
          uint storedData;

          function set(uint x) public {
              storedData = x;
          }

          function get() view public returns (uint retVal) {
              return storedData;
          }
      }
    `
    const {abi, bytecode} = test.compile(source, 'SimpleStorage')
    return burrow.contracts.deploy(abi, bytecode)
      .then((contract: any) => contract.set(42)
        .then(() => contract.get())
      ).then((value) => {
        assert.equal(value, 42)
      })
  }))
})
