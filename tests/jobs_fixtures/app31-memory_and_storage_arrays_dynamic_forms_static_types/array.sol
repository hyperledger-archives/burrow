pragma solidity >=0.0.0;

contract c {

	function intCallWithArray(uint8[4] someUintz) public pure returns (uint){
        return someUintz[3];
	}
	
	function bytesCallWithArray(bytes32[4] someBytez) public pure returns (bytes32) {
		return someBytez[3];
	}
	
	function boolCallWithArray(bool[4] someBoolz) public pure returns (bool){
        return someBoolz[3];
	}

	function addressCallWithArray(address[3] someAddressz) public pure returns (address){
        return someAddressz[2];
	}

	function intMemoryArray() public pure returns (uint8[4]) {
		return [1, 2, 3, 4];
	}

	function bytesMemoryArray() public pure returns (bytes32[5]){
		bytes32[5] memory b;
		b[0] = "hello";
		b[1] = "marmots";
		b[2] = "how";
		b[3] = "are";
		b[4] = "you";
		return b;
	}

	function boolMemoryArray() public pure returns (bool[3]) {
		return [true, false, true];
	}
}
