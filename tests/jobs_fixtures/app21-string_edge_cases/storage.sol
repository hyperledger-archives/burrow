pragma solidity >=0.0.0;

contract SimpleStorage {
  string storedString;

  function setString(string x) {
    storedString = x;
  }

  function getString() constant returns (string retString) {
    return storedString;
  }
}

