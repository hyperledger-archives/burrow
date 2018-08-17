pragma solidity >=0.0.0;

contract GSSingle {
  uint storedData;

  function GSSingle(uint initialValue) {
    storedData = initialValue;
  }

  function set(uint x) {
    storedData = x;
  }

  function get() constant returns (uint retVal) {
    return storedData;
  }
}
