pragma solidity >=0.0.0;

contract base {
  uint storedData;

  function base(uint x) {
  	storedData = 10;
  }

  function set(uint x) {
    storedData = x;
  }

  function get() constant returns (uint retVal) {
    return storedData;
  }
}
