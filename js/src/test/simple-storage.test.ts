import * as assert from 'assert';
import {burrow, compile} from '../test';

describe('Simple storage', function () {
  it('sets and gets a value from a contract', async () => {
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
    const {abi, code} = compile(source, 'SimpleStorage')
    return burrow.contracts.deploy(abi, code)
      .then((contract: any) => contract.set(42)
        .then(() => contract.get())
      ).then((value) => {
        assert.deepStrictEqual(value, [42])
      })
  })
})
