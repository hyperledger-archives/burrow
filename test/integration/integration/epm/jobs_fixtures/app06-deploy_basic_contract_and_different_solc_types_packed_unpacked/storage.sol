pragma solidity >=0.0.0;

contract SimpleStorage {
  bool storedBool;
  bool storedBool2;
  int storedInt;
  uint storedUint;
  address storedAddress;
  bytes32 storedBytes;
  string storedString;

  function setBool(bool x) {
    storedBool = x;
  }

  function getBool() constant returns (bool retBool) {
    return storedBool;
  }

  function setBool2(bool x) {
    storedBool2 = x;
  }

  function getBool2() constant returns (bool retBool) {
    return storedBool2;
  }

  function setInt(int x) {
    storedInt = x;
  }

  function getInt() constant returns (int retInt) {
    return storedInt;
  }

  function setUint(uint x) {
    storedUint = x;
  }

  function getUint() constant returns (uint retUint) {
    return storedUint;
  }

  function setAddress(address x) {
    storedAddress = x;
  }

  function getAddress() constant returns (address retAddress) {
    return storedAddress;
  }

  function setBytes(bytes32 x) {
    storedBytes = x;
  }

  function getBytes() constant returns (bytes32 retBytes) {
    return storedBytes;
  }

  function setString(string x) {
    storedString = x;
  }

  function getString() constant returns (string retString) {
    return storedString;
  }
}

