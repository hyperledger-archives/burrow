import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { burrow } from './test';

describe('Setting and Getting Values:', function () {
  const source = `
pragma solidity >=0.0.0;

contract GetSet {

	uint uintfield;
	bytes32 bytesfield;
	string stringfield;
	bool boolfield;

	function testExist() public pure returns (uint output){
		return 1;
	}

	function setUint(uint input) public {
		uintfield = input;
		return;
	}

	function getUint() public view returns (uint output){
		output = uintfield;
		return output;
	}

	function setBytes(bytes32 input) public {
		bytesfield = input;
		return;
	}

	function getBytes() public view returns (bytes32 output){
		output = bytesfield;
		return output;
	}

	function setString(string memory input) public {
		stringfield = input;
		return;
	}

	function getString() public view returns (string memory output){
		output = stringfield;
		return output;
	}

	function setBool(bool input) public {
		boolfield = input;
		return;
	}

	function getBool() public view returns (bool output){
		output = boolfield;
		return output;
	}
}
`;

  const testUint = 42;
  const testBytes = 'DEADBEEF00000000000000000000000000000000000000000000000000000000';
  const testString = 'Hello World!';
  const testBool = true;

  let TestContract: any;

  before(async () => {
    const contract = compile(source, 'GetSet');
    TestContract = await contract.deploy(burrow);
  });

  it('Uint', async () => {
    await TestContract.setUint(testUint);
    const output = await TestContract.getUint();
    assert.strictEqual(output[0], testUint);
  });

  it('Bool', async () => {
    await TestContract.setBool(testBool);
    const output = await TestContract.getBool();
    assert.strictEqual(output[0], testBool);
  });

  it('Bytes', async () => {
    await TestContract.setBytes(testBytes);
    const output = await TestContract.getBytes();
    assert.strictEqual(output[0], testBytes);
  });

  it('String', async () => {
    await TestContract.setString(testString);
    const output = await TestContract.getString();
    assert.strictEqual(output[0], testString);
  });
});
