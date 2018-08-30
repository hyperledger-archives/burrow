pragma solidity >=0.0.0;

contract GSSingle {
  uint storedData;

  constructor(uint initialValue) public {
    storedData = initialValue;
  }

  function set(uint x) public {
    storedData = x;
  }

  function get() public constant returns (uint retVal) {
    return storedData;
  }
}
