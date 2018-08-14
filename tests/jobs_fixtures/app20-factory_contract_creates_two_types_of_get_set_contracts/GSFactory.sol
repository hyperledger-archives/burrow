pragma solidity >=0.0.0;

contract GSContract {
  uint storedData;

  function set(uint x) {
    storedData = x;
  }

  function get() constant returns (uint retVal) {
    return storedData;
  }
}

contract GSContract2 {
  uint storedData;

  function set2(uint x) {
    storedData = x;
  }

  function get2() constant returns (uint retVal) {
    return storedData;
  }
}


contract GSFactory {
	address lastCreated;
	function create() returns (address GSAddr) {
		lastCreated = new GSContract();
		return lastCreated;
	}

	function create2() returns (address GSAddr) {
		lastCreated = new GSContract2();
		return lastCreated;
	}

	function getLast() returns (address GSAddr) {
		return lastCreated;
	}
}