pragma solidity >=0.0.0;

contract SimpleStorage {
  bytes storedString;

  function setString(bytes x) {
    storedString = x;
  }

  function getString() constant returns (bytes retString) {
    return storedString;
  }
}

