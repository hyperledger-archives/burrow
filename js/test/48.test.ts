import * as assert from 'assert';
import { compile } from '../src/contracts/compile'
import { burrow } from './test';

describe('#48', function () {it('#48', async () => {
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
    `
    const {abi, code} = compile(source, 'Test')
    return burrow.contracts.deploy(abi, code)
      .then((contract: any) => contract.getCombination())
      .then(([number, address]) => {
        assert.strictEqual(number, 100)
        assert.strictEqual(address.length, 40)
      })
  })
})
