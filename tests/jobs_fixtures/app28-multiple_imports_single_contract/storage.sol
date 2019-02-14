pragma solidity >=0.0.0;

contract SimpleStorage {
  string storedString1;
  string storedString2;

  function setString(string memory x1, string memory x2) public {
    storedString2 = x2;
    storedString1 = x1;
  }

  function getString1() public view returns (string memory retString1) {
    return storedString1;
  }

  function getString2() public view returns (string memory retString2) {
    return storedString2;
  }
}

