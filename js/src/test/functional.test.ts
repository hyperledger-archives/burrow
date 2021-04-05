import * as assert from 'assert';
import {burrow, compile} from '../test';
import { Contract } from '..';

describe('Functional Contract Usage', function () {it('#Constructor usage', async () => {
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
    `
    const {abi, code} = compile(source, 'Test')
    const contract: any = new Contract(abi, code, null, burrow)

    let A1
    let A2

    assert.strictEqual(contract.address, null)

    // Use the _contructor method to creat two contracts
    return Promise.all(
      [contract._constructor('88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F'),
        contract._constructor('ABCDEFABCDEFABCDEFABCDEFABCDEFABCDEF0123')])
      .then(([address1, address2]) => {
        assert.strictEqual(contract.address, null)
        A1 = address1
        A2 = address2

        return Promise.all(
          [contract.getCombination.at(A1),
            contract.getCombination.at(A2)])
      })
      .then(([object1, object2]) => {
        const expected1 = [100, A1, 'hello moto', '000000000000000000000000000000000000000000000000DEADBEEFFEEDFACE', '88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F']
        const expected2 = [100, A2, 'hello moto', '000000000000000000000000000000000000000000000000DEADBEEFFEEDFACE', 'ABCDEFABCDEFABCDEFABCDEFABCDEFABCDEF0123']
        assert.deepEqual(object1, expected1)
        assert.deepEqual(object2, expected2)
      })
  })
})
