import * as assert from 'assert';
import { compile } from '../src/contracts/compile'
import { burrow } from './test';

describe('#50', function () {it('#50', async () => {
    const source = `
      pragma solidity >=0.0.0;
      contract SimpleStorage {
          uint storedData;

          function set(uint x) public {
              storedData = x;
          }

          function get() public view returns (uint retVal) {
              return storedData;
          }
      }
    `
    const {abi, code} = compile(source, 'SimpleStorage')
    return burrow.contracts.deploy(abi, code)
      .then((contract: any) => contract.set(42)
        .then(() => contract.get.call())
      ).then((value) => {
        assert.deepStrictEqual(value, [42])
      })
  })
})
