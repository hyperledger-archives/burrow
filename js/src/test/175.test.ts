import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { ContractInstance, getMetadata } from '../contracts/contract';
import { burrow } from './test';

describe('#175', function () {
  it('#175', async () => {
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
    `;

    const contract = compile<ContractInstance>(source, 'Contract');

    const instance1 = (await contract.deploy(burrow, 'contract1')) as any;
    const instance2 = await contract.deploy(burrow, 'contract2');

    const address = getMetadata(instance2).address;
    const ps = await Promise.all([
      // Note using the default address from the deploy
      instance1.getName(),
      // Using the .at() to specify the second deployed contract
      instance1.getName.at(address)(),
    ]);
    const [[name1], [name2]] = ps;
    assert.strictEqual(name1, 'contract1');
    assert.strictEqual(name2, 'contract2');
  });
});
