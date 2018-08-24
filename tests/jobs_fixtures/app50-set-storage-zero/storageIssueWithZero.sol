pragma solidity ^0.4.20;

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

  function getInt() constant public returns (int retInt) {
    return storedInt;
  }

  function setUint(uint x) public {
    storedUint = x;
  }

  function setUintToZero() public {
    storedUint = 0;
  }

  function getUint() constant public returns (uint retUint) {
    return storedUint;
  }

}
