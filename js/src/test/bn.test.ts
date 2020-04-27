import * as assert from 'assert';
import { BigNumber } from 'ethers/lib/ethers';
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
    const expected = BigNumber.from('10000000000000000000');
    assert.ok(expected.eq(number));

    await instance.mul(100, -300).then(([number]: any) => {
      assert.strictEqual(number, -30000);
    });

    await instance.mul(BigNumber.from('18446744073709551616'), 102).then(([number]: any) => {
      assert.ok(BigNumber.from('1881567895518374264832').eq(number));
    });
  });
});
