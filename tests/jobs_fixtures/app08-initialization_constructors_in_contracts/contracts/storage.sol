pragma solidity >=0.0.0;

contract SimpleConstructorInt {
  uint public storedData;

  function SimpleConstructorInt(uint x, uint y) {
    storedData = x;
  }
}

contract SimpleConstructorBool {
  bool public storedData;

  function SimpleConstructorBool(bool x, bool y) {
    storedData = x;
  }
}

contract SimpleConstructorString {
  string public storedData;

  function SimpleConstructorString(string x, string y) {
    storedData = x;
  }
}

contract SimpleConstructorBytes {
  bytes32 public storedData;

  function SimpleConstructorBytes(bytes32 x, bytes32 y) {
    storedData = x;
  }
}

contract SimpleConstructorArray {
  uint[3] public storedData;

  function SimpleConstructorArray(uint[3] x, uint[3] y) {
    storedData = x;
  }

  function get() returns (uint[3]) {
    return storedData;
  }
}