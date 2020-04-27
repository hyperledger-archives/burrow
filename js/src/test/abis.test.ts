import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { getAddress } from '../contracts/contract';
import { burrow } from './test';

describe('Abi', function () {
  const source = `
pragma solidity >=0.0.0;

contract random {
	function getRandomNumber() public pure returns (uint) {
		return 55;
	}
}
  `;
  // TODO: understand why abi storage isn't working
  it('Call contract via burrow side Abi', async () => {
    const contract = compile(source, 'random');
    const instance: any = await contract.deploy(burrow);
    await burrow.namereg.set('random', getAddress(instance));
    const entry = await burrow.namereg.get('random');
    const address = entry.getData();
    console.log(address);
    const contractOut: any = await contract.at(address, burrow);
    const number = await contractOut.getRandomNumber();
    assert.strictEqual(number[0], 55);
  });
});
