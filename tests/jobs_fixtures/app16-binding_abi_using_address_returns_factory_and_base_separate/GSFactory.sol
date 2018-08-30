pragma solidity >=0.0.0;

import "./GSContract.sol";

contract GSFactory {
	address lastCreated;
	function create() public returns (address GSAddr) {
		lastCreated = new GSContract();
		return lastCreated;
	}

	function getLast() public view returns (address GSAddr) {
		return lastCreated;
	}
}