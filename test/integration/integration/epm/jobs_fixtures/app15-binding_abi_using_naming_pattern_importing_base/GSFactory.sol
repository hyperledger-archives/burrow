pragma solidity >=0.0.0;

import "./GSContract.sol";

contract GSFactory {
	address lastCreated;
	function create() returns (address GSAddr) {
		lastCreated = new GSContract();
		return lastCreated;
	}

	function getLast() returns (address GSAddr) {
		return lastCreated;
	}
}