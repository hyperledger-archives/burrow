import * as assert from 'assert';
import {burrow, compile} from '../test';

describe('#61', function () {it('#61', async () => {
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
    `
    const {abi, code} = compile(source, 'SimpleStorage')
    return burrow.contracts.deploy(abi, code, null, '88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F')
      .then((contract: any) => contract.get())
      .then((value) => {
        assert.deepStrictEqual(value, ['88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F'])
      })
  })
})
