import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { client } from './test';

describe('#46', function () {
  it('#46', async () => {
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
    `;

    const contract = compile(source, 'Test');
    return contract
      .deploy(client)
      .then((instance: any) =>
        instance.setName('Batman').then(() => Promise.all([instance.getNameConstant(), instance.getName()])),
      )
      .then(([constant, nonConstant]) => {
        assert.strictEqual(constant[0], nonConstant[0]);
      });
  });
});
