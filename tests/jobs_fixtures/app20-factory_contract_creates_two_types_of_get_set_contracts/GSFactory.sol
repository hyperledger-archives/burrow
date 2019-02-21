pragma solidity >=0.0.0;

contract GSContract {
  uint storedData;

  function set(uint x) public {
    storedData = x;
  }

  function get() public view returns (uint retVal) {
    return storedData;
  }
}

contract GSContract2 {
  uint storedData;

  function set2(uint x) public {
    storedData = x;
  }

  function get2() public view returns (uint retVal) {
    return storedData;
  }
}


contract GSFactory {
	address lastCreated;
	function create() public returns (address GSAddr) {
		lastCreated = address(new GSContract());
		return lastCreated;
	}

	function create2() public returns (address GSAddr) {
		lastCreated = address(new GSContract2());
		return lastCreated;
	}

	function getLast() public view returns (address GSAddr) {
		return lastCreated;
	}
}