import * as assert from 'assert';
import {burrow, compile} from "../test";

describe('Multiple return types', function () {
  it('#42', async () => {
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
    `
    const {abi, code} = compile(source, 'Test')
    const contract: any = await burrow.contracts.deploy(abi, code, {
      call: function (result) {
        return {values: result.values, raw: result.raw}
      }
    })
    const expected = {
      values: {
        _number: 100,
        _address: contract.address,
        _saying: 'hello moto',
        _randomBytes: '000000000000000000000000000000000000000000000000DEADBEEFFEEDFACE'
      },
      raw: [100, contract.address, 'hello moto', '000000000000000000000000000000000000000000000000DEADBEEFFEEDFACE']
    }
    const result = await contract.getCombination();
    assert.deepStrictEqual(result, expected)
  })
})
