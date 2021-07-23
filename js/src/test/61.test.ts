import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { client } from './test';

describe('#61', function () {
  it('#61', async () => {
    const source = `
      pragma solidity >=0.0.0;
      contract SimpleStorage {
          address storedData;

          constructor(address x) public {
              storedData = x;
          }

          function get() public view returns (address retVal) {
              return storedData;
          }
      }
    `;
    const contract = compile(source, 'SimpleStorage');
    const instance = await contract.deploy(client, '88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F');
    const value = await instance.get();
    assert.deepStrictEqual([...value], ['88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F']);
  });
});
