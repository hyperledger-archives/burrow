import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { burrow } from './test';

describe('#44', function () {
  it('#44', async () => {
    const source = `
      pragma solidity >=0.0.0;
      contract SimpleStorage {
          address storedData;

          function set(address x) public {
              storedData = x;
          }

          function get() public view returns (address retVal) {
              return storedData;
          }
      }
    `;
    const contract = compile(source, 'SimpleStorage');
    const instance: any = await contract.deploy(burrow);
    await instance.set('88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F');

    const data = await instance.get();
    assert.deepStrictEqual([...data], ['88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F']);
  });
});
