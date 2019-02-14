pragma solidity >=0.0.0;

contract SimpleStorage {
  int storedData;

  function set(int x) public  {
    storedData = x;
  }

  function get() public view returns (int /* retVal */) {
    return storedData;
  }
}