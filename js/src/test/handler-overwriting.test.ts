import * as assert from 'assert';
import { CallResult } from '../contracts/call';
import { compile } from '../contracts/compile';
import { getMetadata } from '../contracts/contract';
import { withoutArrayElements } from '../convert';
import { burrow } from './test';

describe('Testing Per-contract handler overwriting', function () {
  // {handlers: {call: function (result) { return {super: result.values, man: result.raw} }}})

  it('#17 Testing Per-contract handler overwriting', async () => {
    const source = `
      pragma solidity >=0.0.0;
      contract Test {

          function getAddress() public view returns (address) {
            return address(this);
          }

          function getNumber() public pure returns (uint) {
            return 100;
          }

          function getCombination() public view returns (uint _number, address _address, string memory _saying, bytes32 _randomBytes) {
            _number = 100;
            _address = address(this);
            _saying = "hello moto";
            _randomBytes = bytes32(uint256(0xDEADBEEFFEEDFACE));
          }

      }
    `;

    const instance: any = await compile(source, 'Test').deployWith(burrow, {
      handler: function ({ result }: CallResult) {
        return {
          values: withoutArrayElements(result),
          raw: [...result],
        };
      },
    });
    const address = getMetadata(instance).address;
    const returnObject = await instance.getCombination();
    const expected = {
      values: {
        _number: 100,
        _address: address,
        _saying: 'hello moto',
        _randomBytes: '000000000000000000000000000000000000000000000000DEADBEEFFEEDFACE',
      },
      raw: [100, address, 'hello moto', '000000000000000000000000000000000000000000000000DEADBEEFFEEDFACE'],
    };
    assert.deepStrictEqual(returnObject, expected);
  });
});
