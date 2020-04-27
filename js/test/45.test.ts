import * as assert from 'assert';
import { compile } from '../src/contracts/compile'
import { burrow } from './test';

describe('#45', function () {

  it('nottherealbatman', async () => {
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
    `
    const {abi, code} = compile(source, 'Test')
    return burrow.contracts.deploy(abi, code).then((contract: any) =>
      contract.setName('Batman')
        .then(() => contract.getName())
    ).then((value) => {
      assert.deepStrictEqual(value, ['Batman'])
    })
  })

  it('rguikers', async () => {
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
    `

    const {abi, code} = compile(source, 'Test')
    return burrow.contracts.deploy(abi, code).then((contract: any) =>
      Promise.all([contract.getAddress(), contract.getNumber()])
        .then(([address, number]) => {
          assert.strictEqual(address[0].length, 40)
          assert.strictEqual(number[0], 100)
        })
    )
  })
})
