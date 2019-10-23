pragma solidity ^0.5;

contract storageIssueWithZero {

  int private storedInt;
  uint private storedUint;
  int foo = 102;

  function setInt(int x) public {
    storedInt = x;
  }

  function setIntToZero() public {
    storedInt = 0;
  }

  function getInt() view public returns (int retInt) {
    return storedInt;
  }

  function setUint(uint x) public {
    storedUint = x;
  }

  function setUintToZero() public {
    storedUint = 0;
  }

  function getUint() view public returns (uint retUint) {
    return storedUint;
  }

}
