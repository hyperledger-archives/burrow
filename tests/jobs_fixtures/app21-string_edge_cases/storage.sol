pragma solidity >=0.0.0;

contract SimpleStorage {
  string storedString;

  function setString(string x) public {
    storedString = x;
  }

  function getString() public constant returns (string retString) {
    return storedString;
  }
}

