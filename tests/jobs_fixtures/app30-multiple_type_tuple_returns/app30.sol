pragma solidity >=0.0.0;

contract multiReturn {

	function getInts() public pure returns (uint, int) {
  		return (1, 2);
	}
	function getBools() public pure returns (bool, bool) {
		return (true, false);
	}
	function getBytes() public pure returns (bytes32, bytes32, bytes32, bytes32) {
		return ("Hello", "World", "of", "marmots");
	}
	function getInterMixed() public view
		returns (
			address myAddress,
			bytes2 elaborate,
			uint8 funNumber
		) {
		myAddress = address(this);
		elaborate = "is";
		funNumber = 1;
	}
}

