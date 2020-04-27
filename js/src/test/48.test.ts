import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { burrow } from './test';

describe('#48', function () {
  it('#48', async () => {
    const source = `
      pragma solidity >=0.0.0;
      contract Test {

          function getAddress() public view returns (address) {
            return address(this);
          }

          function getNumber() public pure returns (uint) {
            return 100;
          }

          function getCombination() public view returns (uint _number, address _address) {
            _number = 100;
            _address = address(this);
          }

      }
    `;
    const contract = compile(source, 'Test');
    return contract
      .deploy(burrow)
      .then((instance: any) => instance.getCombination())
      .then(([number, address]) => {
        assert.strictEqual(number, 100);
        assert.strictEqual(address.length, 40);
      });
  });
});
