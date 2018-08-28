pragma solidity >=0.0.0;

contract Storage {
  uint storedData;

  function set(uint x) public {
    storedData = x;
  }

  function get() public constant returns (uint retVal) {
    return storedData;
  }
}

contract GSFactory {
  address lastCreated;

  function create() public returns (address GSAddr) {
    lastCreated = new Storage();
    return lastCreated;
  }

  function last() public view returns (address GSAddr) {
    return lastCreated;
  }
}
