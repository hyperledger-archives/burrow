import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { burrow } from './test';

describe('#50', function () {
  it('#50', async () => {
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
    `;
    const contract = compile(source, 'SimpleStorage');
    const instance = await contract.deploy(burrow);
    await instance.set(42);
    const value = await instance.get.call();
    assert.deepStrictEqual([...value], [42]);
  });
});
