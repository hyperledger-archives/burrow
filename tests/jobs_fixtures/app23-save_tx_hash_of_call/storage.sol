pragma solidity >=0.0.0;

contract Storage {
  int storedData;

  function set(int x) public {
    storedData = x;
  }

  function get() public constant returns (int retVal) {
    return storedData;
  }
}

