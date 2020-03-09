pragma solidity ^0.4.21;

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

	function getUint() public  constant returns (uint output){
		output = uintfield;
		return;
	}

	function setBytes(bytes32 input) public {
		bytesfield = input;
		return;
	}

	function getBytes() public  constant returns (bytes32 output){
		output = bytesfield;
		return;
	}

	function setString(string input) public {
		stringfield = input;
		return;
	}

	function getString() public constant returns (string output){
		output = stringfield;
		return;
	}

	function setBool(bool input) public {
		boolfield = input;
		return;
	}

	function getBool() public constant returns (bool output){
		output = boolfield;
		return;
	}

}