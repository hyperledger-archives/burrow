import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { burrow } from './test';

describe('BN', function () {
  it('BN', async () => {
    const source = `
      pragma solidity >=0.0.0;
      contract Test {
          function mul(int a, int b) public pure returns (int) {
            return a * b;
          }

          function getNumber() public pure returns (uint) {
            return 1e19;
          }
      }
    `;
    const contract = compile(source, 'Test');
    const instance = await contract.deploy(burrow);

    const [number] = await instance.getNumber();
    assert.strictEqual(number.toString(), '10000000000000000000');

    const [smallNumber] = await instance.mul(100, -300);
    assert.strictEqual(smallNumber, -30000);

    const [bigNumber] = await instance.mul('18446744073709551616', 102);
    assert.strictEqual(bigNumber.toString(), '1881567895518374264832');
  });
});
