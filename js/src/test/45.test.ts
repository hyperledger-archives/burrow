import * as assert from 'assert';
import * as test from '../test';

const Test = test.Test();

describe('#45', function () {
  before(Test.before())
  after(Test.after())

  this.timeout(10 * 1000)

  it('nottherealbatman', Test.it(function (burrow) {
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
    const {abi, bytecode} = test.compile(source, 'Test')
    return burrow.contracts.deploy(abi, bytecode).then((contract: any) =>
      contract.setName('Batman')
        .then(() => contract.getName())
    ).then((value) => {
      assert.equal(value, 'Batman')
    })
  }))

  it('rguikers', Test.it(function (burrow) {
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

    const {abi, bytecode} = test.compile(source, 'Test')
    return burrow.contracts.deploy(abi, bytecode).then((contract: any) =>
      Promise.all([contract.getAddress(), contract.getNumber()])
        .then(([address, number]) => {
          assert.equal(address[0].length, 40)
          assert.equal(number[0], 100)
        })
    )
  }))
})
