import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { getMetadata } from '../contracts/contract';
import { burrow } from './test';

describe('Functional Contract Usage', function () {
  it('#Constructor usage', async () => {
    const source = `
      pragma solidity >=0.0.0;
      contract Test {
        address storedData;
        constructor(address x) public {
          storedData = x;
        }

        function getAddress() public view returns (address) {
          return address(this);
        }

        function getNumber() public pure returns (uint) {
          return 100;
        }

        function getCombination() public view returns (uint _number, address _address, string memory _saying, bytes32 _randomBytes, address _stored) {
          _number = 100;
          _address = address(this);
          _saying = "hello moto";
          _randomBytes = bytes32(uint256(0xDEADBEEFFEEDFACE));
          _stored = storedData;
        }

      }
    `;
    const contract = compile(source, 'Test');

    const [instance1, instance2] = await Promise.all([
      contract.deploy(burrow, '88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F'),
      contract.deploy(burrow, 'ABCDEFABCDEFABCDEFABCDEFABCDEFABCDEF0123'),
    ]);

    const address1 = getMetadata(instance1).address;
    const address2 = getMetadata(instance2).address;

    const [ret1, ret2] = await Promise.all([
      instance1.getCombination.at(address1)(),
      instance1.getCombination.at(address2)(),
    ]);

    const expected1 = [
      100,
      address1,
      'hello moto',
      '000000000000000000000000000000000000000000000000DEADBEEFFEEDFACE',
      '88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F',
    ];
    const expected2 = [
      100,
      address2,
      'hello moto',
      '000000000000000000000000000000000000000000000000DEADBEEFFEEDFACE',
      'ABCDEFABCDEFABCDEFABCDEFABCDEFABCDEF0123',
    ];
    assert.deepStrictEqual([...ret1], expected1);
    assert.deepStrictEqual([...ret2], expected2);
    // Check we are assigning names to hybrid record/array Result object
    assert.strictEqual(ret2._saying, 'hello moto');
  });
});
