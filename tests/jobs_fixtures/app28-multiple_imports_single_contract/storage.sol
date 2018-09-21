pragma solidity >=0.0.0;

contract SimpleStorage {
  string storedString1;
  string storedString2;

  function setString(string x1, string x2) public {
    storedString2 = x2;
    storedString1 = x1;
  }

  function getString1() public constant returns (string retString1) {
    return storedString1;
  }

  function getString2() public constant returns (string retString2) {
    return storedString2;
  }
}

