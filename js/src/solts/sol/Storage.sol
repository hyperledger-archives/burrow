pragma solidity >=0.0.0;

contract Storage {
  int data;

  constructor(int x) public {
      data = x;
  }

  function set(int x) public {
    data = x;
  }

  function get() public view returns (int ret) {
    return data;
  }
}

