pragma solidity >=0.0.0;

contract base {
  uint storedData;

  constructor(uint /* x */) public {
  	storedData = 10;
  }

  function set(uint x) public {
    storedData = x;
  }

  function get() view public returns (uint retVal) {
    return storedData;
  }
}
