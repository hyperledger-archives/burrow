pragma solidity >=0.0.0;

contract binary {
  uint storedData;

  function set(uint x) public {
    storedData = x;
  }

  function get() public view returns (uint retVal) {
    return storedData;
  }
}

