pragma solidity >=0.0.0;

contract SimpleConstructorInt {
  uint public storedData;

  constructor(uint x, uint /* y */) public {
    storedData = x;
  }
}

contract SimpleConstructorBool {
  bool public storedData;

  constructor(bool x, bool /* y */) public {
    storedData = x;
  }
}

contract SimpleConstructorString {
  string public storedData;

  constructor(string memory x, string memory /* y */) public {
    storedData = x;
  }
}

contract SimpleConstructorBytes {
  bytes32 public storedData;

  constructor(bytes32 x, bytes32 /* y */) public {
    storedData = x;
  }
}

contract SimpleConstructorArray {
  uint[3] public storedData;

  constructor(uint[3] memory x, uint[3] memory /* y */) public {
    storedData = x;
  }

  function get() public view returns (uint[3] memory) {
    return storedData;
  }
}