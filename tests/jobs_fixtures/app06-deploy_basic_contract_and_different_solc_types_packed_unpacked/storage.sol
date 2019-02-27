pragma solidity >=0.0.0;

contract SimpleStorage {
  bool storedBool;
  bool storedBool2;
  int storedInt;
  uint storedUint;
  address storedAddress;
  bytes32 storedBytes;
  string storedString;

  function setBool(bool x) public {
    storedBool = x;
  }

  function getBool() view public returns (bool retBool) {
    return storedBool;
  }

  function setBool2(bool x) public {
    storedBool2 = x;
  }

  function getBool2() view public returns (bool retBool) {
    return storedBool2;
  }

  function setInt(int x) public {
    storedInt = x;
  }

  function getInt() view public returns (int retInt) {
    return storedInt;
  }

  function setUint(uint x) public {
    storedUint = x;
  }

  function getUint() view public returns (uint retUint) {
    return storedUint;
  }

  function setAddress(address x) public {
    storedAddress = x;
  }

  function getAddress() view public returns (address retAddress) {
    return storedAddress;
  }

  function setBytes(bytes32 x) public {
    storedBytes = x;
  }

  function getBytes() view public returns (bytes32 retBytes) {
    return storedBytes;
  }

  function setString(string memory x) public {
    storedString = x;
  }

  function getString() view public returns (string memory retString) {
    return storedString;
  }
}