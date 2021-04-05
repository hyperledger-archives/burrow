import * as assert from 'assert';
import {burrow, compile} from "../test";

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
    `
    const {abi, code} = compile(source, 'SimpleStorage');
    const contract: any = await burrow.contracts.deploy(abi, code);
    await contract.set('88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F');

    const data = await contract.get();
    assert.deepStrictEqual(data,[ '88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F']);
  })
})
