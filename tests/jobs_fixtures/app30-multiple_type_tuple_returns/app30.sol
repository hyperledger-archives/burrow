pragma solidity >=0.0.0;

contract multiReturn {

	function getInts() returns (uint, int) {
  		return (1, 2);
	}
	function getBools() returns (bool, bool) {
		return (true, false);
	}
	function getBytes() returns (bytes32, bytes32, bytes32, bytes32) {
		return ("Hello", "World", "of", "marmots");
	}
	function getInterMixed() 
		returns (
			address myAddress,
			bytes2 elaborate,
			uint8 funNumber
		) {
		myAddress = this;
		elaborate = "is";
		funNumber = 1;
	}
}

