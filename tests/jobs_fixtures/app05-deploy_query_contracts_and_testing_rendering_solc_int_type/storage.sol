pragma solidity >=0.0.0;

contract Storage {
  int storedData;

  function set(int x) {
    storedData = x;
  }

  function get() constant returns (int retVal) {
    return storedData;
  }
}

