pragma solidity >=0.0.0;

contract Storage {
  uint storedData;

  function set(uint x) {
    storedData = x;
  }

  function get() constant returns (uint retVal) {
    return storedData;
  }
}

contract GSFactory {
  address lastCreated;

  function create() returns (address GSAddr) {
    lastCreated = new Storage();
    return lastCreated;
  }

  function last() returns (address GSAddr) {
    return lastCreated;
  }
}
