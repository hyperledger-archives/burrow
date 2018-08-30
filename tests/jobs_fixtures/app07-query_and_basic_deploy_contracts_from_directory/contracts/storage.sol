pragma solidity >=0.0.0;

contract SimpleStorage {
  uint storedData;

  function set(uint x) public {
    storedData = x;
  }

  function get() constant public returns (uint retVal) {
    return storedData;
  }
}

