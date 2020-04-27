import * as assert from 'assert';
import { compile } from '../src/contracts/compile'
import { burrow } from './test';

describe('#175', function () {it('#175', async () => {
    const source = `
    pragma solidity >=0.0.0;
      contract Contract {
        string thename;
        constructor(string memory newName) public {
          thename = newName;
        }
        function getName() public view returns (string memory name) {
          return thename;
        }
      }
    `
    let contract
    let A2

    const {abi, code} = compile(source, 'Contract')
    return burrow.contracts.deploy(abi, code, null, 'contract1').then((C) => {
      contract = C
      return contract._constructor('contract2')
    }).then((address) => {
      A2 = address
      return Promise.all(
        [contract.getName(),      // Note using the default address from the deploy
          contract.getName.at(A2)])   // Using the .at() to specify the second deployed contract
    }).then(([result1, result2]) => {
      assert.strictEqual(result1[0], 'contract1')
      assert.strictEqual(result2[0], 'contract2')
    })
  })
})
