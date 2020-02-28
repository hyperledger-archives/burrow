import * as assert from 'assert';
import * as test from '../test';

const Test = test.Test();

describe('#42', function () {
  before(Test.before())
  after(Test.after())

  this.timeout(10 * 1000)

  it('#42', Test.it(function (burrow) {
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
    let address: string;
    const {abi, bytecode} = test.compile(source, 'Test')
    return burrow.contracts.deploy(abi, bytecode, {call: function (result) { return {values: result.values, raw: result.raw} }}).then((contract: any) => {
      address = contract.address
      return contract.getCombination()
    }).then((returnObject) => {
      const expected = {
        values: {
          _number: 100,
          _address: address,
          _saying: 'hello moto',
          _randomBytes: '000000000000000000000000000000000000000000000000DEADBEEFFEEDFACE'
        },
        raw: [100, address, 'hello moto', '000000000000000000000000000000000000000000000000DEADBEEFFEEDFACE']
      }
      assert.deepEqual(returnObject, expected)
    })
  }))
})
