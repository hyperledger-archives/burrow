pragma solidity ^0.4.24;

contract SimpleStorage {
  int storedData;

  function set(int x) public  {
    storedData = x;
  }

  function get() public constant returns (int /* retVal */) {
    return storedData;
  }
}