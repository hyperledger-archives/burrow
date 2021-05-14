import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { burrow } from './test';

describe('#47', function () {
  it('#47', async () => {
    const source = `
      pragma solidity >=0.0.0;
      contract Test{
        string _withSpace = "  Pieter";
        string _withoutSpace = "Pieter";

        function getWithSpaceConstant() public view returns (string memory) {
          return _withSpace;
        }

        function getWithoutSpaceConstant () public view returns (string memory) {
          return _withoutSpace;
        }
      }
    `;
    const contract = compile(source, 'Test');
    return contract
      .deploy(burrow)
      .then((instance: any) => Promise.all([instance.getWithSpaceConstant(), instance.getWithoutSpaceConstant()]))
      .then(([withSpace, withoutSpace]) => {
        assert.deepStrictEqual(withSpace, ['  Pieter']);
        assert.deepStrictEqual(withoutSpace, ['Pieter']);
      });
  });
});
