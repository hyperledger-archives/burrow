import * as assert from 'assert';
import * as test from '../test';

const Test = test.Test();

describe('#46', function () {
  before(Test.before())
  after(Test.after())

  this.timeout(10 * 1000)

  it('#46', Test.it(function (burrow) {
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

    const {abi, bytecode} = test.compile(source, 'Test')
    return burrow.contracts.deploy(abi, bytecode)
      .then((contract: any) => contract.setName('Batman')
        .then(() => Promise.all([contract.getNameConstant(), contract.getName()])))
      .then(([constant, nonConstant]) => {
        assert.equal(constant[0], nonConstant[0])
      })
  }))
})
