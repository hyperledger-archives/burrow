import * as assert from 'assert';
import {burrow, compile} from "../test";

describe('Abi', function () {

  const source = `
pragma solidity >=0.0.0;

contract random {
	function getRandomNumber() public pure returns (uint) {
		return 55;
	}
}
  `
  // TODO: understand why abi storage isn't working
  it('Call contract via burrow side Abi', async () => {
    const {abi, code} = compile(source, 'random')
    const contractIn: any = await burrow.contracts.deploy(abi, code)
    await burrow.namereg.set('random', contractIn.address)
    const entry = await burrow.namereg.get('random')
    const address = entry.getData();
    console.log(address)
    const contractOut: any = await burrow.contracts.fromAddress(address)
    const number = await contractOut.getRandomNumber()
    assert.strictEqual(number[0], 55)
  })
})
