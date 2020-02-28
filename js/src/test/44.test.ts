import * as assert from 'assert';
import * as test from '../test';

const Test = test.Test();

describe('#44', function () {
  before(Test.before())
  after(Test.after())

  this.timeout(60 * 1000)

  it('#44', Test.it(async (burrow) => {
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
    const {abi, bytecode} = test.compile(source, 'SimpleStorage');
    const contract: any = await burrow.contracts.deploy(abi, bytecode);
    await contract.set('88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F');

    const data = await contract.get();
    assert.equal(data, '88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F');
  }))
})
