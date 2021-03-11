import * as assert from 'assert';
import { burrow, compile } from '../test';
import BN from 'bn.js';

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
    `
        const { abi, code } = compile(source, 'Test')
        let contract: any = await burrow.contracts.deploy(abi, code);

        await contract.getNumber()
            .then(([number]) => {
                assert.strictEqual(new BN('10000000000000000000').cmp(number), 0)
            })

        await contract.mul(100, -300)
            .then(([number]) => {
                assert.strictEqual(number, -30000)
            })

        await contract.mul(new BN('18446744073709551616'), 102)
            .then(([number]) => {
                assert.strictEqual(new BN('1881567895518374264832').cmp(number), 0)
            })
    })
})
