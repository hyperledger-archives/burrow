import * as assert from 'assert';
import {burrow, compile} from '../test';

describe('#46', function () {it('#46', async () => {
    const source = `
      pragma solidity >=0.0.0;
      contract Test{

        string _name;

        function setName(string memory newname) public {
          _name = newname;
        }

        function getNameConstant() public view returns (string memory) {
          return _name;
        }

        function getName() public view returns (string memory) {
          return _name;
        }
      }
    `

    const {abi, code} = compile(source, 'Test')
    return burrow.contracts.deploy(abi, code)
      .then((contract: any) => contract.setName('Batman')
        .then(() => Promise.all([contract.getNameConstant(), contract.getName()])))
      .then(([constant, nonConstant]) => {
        assert.strictEqual(constant[0], nonConstant[0])
      })
  })
})
