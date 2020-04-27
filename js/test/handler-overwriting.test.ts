import * as assert from 'assert';
import {burrow, compile} from "./test";


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
    `

    const {abi, code} = compile(source, 'Test')
    const contract: any = await burrow.contracts.deploy(abi, code, {
      call: function (result_1) {
        return {values: result_1.values, raw: result_1.raw};
      }
    });
    const address = contract.address;
    const returnObject = await contract.getCombination();
    const expected = {
      values: {
        _number: 100,
        _address: address,
        _saying: 'hello moto',
        _randomBytes: '000000000000000000000000000000000000000000000000DEADBEEFFEEDFACE'
      },
      raw: [100, address, 'hello moto', '000000000000000000000000000000000000000000000000DEADBEEFFEEDFACE']
    };
    assert.deepStrictEqual(returnObject, expected);
  })
})
