import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { client } from './test';

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
    `;
    const contract = compile(source, 'SimpleStorage');
    return contract
      .deploy(client)
      .then((instance: any) => instance.set(42).then(() => instance.get()))
      .then((value) => {
        assert.deepStrictEqual([...value], [42]);
      });
  });
});
