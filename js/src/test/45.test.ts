import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { burrow } from './test';

describe('#45', function () {
  it('Set/get memory string', async () => {
    const source = `
      pragma solidity >=0.0.0;
      contract Test {
          string _name;

          function add(int a, int b) public pure returns (int sum) {
              sum = a + b;
          }

          function setName(string memory newname) public {
             _name = newname;
          }

          function getName() public view returns (string memory) {
              return _name;
          }
      }
    `;
    const contract = compile(source, 'Test');
    return contract
      .deploy(burrow)
      .then((instance: any) => instance.setName('Batman').then(() => instance.getName()))
      .then((value) => {
        assert.deepStrictEqual(value, ['Batman']);
      });
  });

  it('get number/address', async () => {
    const source = `
      pragma solidity >=0.0.0;
      contract Test {

          function getAddress() public view returns (address) {
            return address(this);
          }

          function getNumber() public pure returns (uint) {
            return 100;
          }

      }
    `;

    const contract = compile(source, 'Test');
    return contract.deploy(burrow).then((instance: any) =>
      Promise.all([instance.getAddress(), instance.getNumber()]).then(([address, number]) => {
        assert.strictEqual(address[0].length, 40);
        assert.strictEqual(number[0], 100);
      }),
    );
  });
});
