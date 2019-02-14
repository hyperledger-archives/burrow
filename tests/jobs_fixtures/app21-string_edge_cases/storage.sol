pragma solidity >=0.0.0;

contract SimpleStorage {
  string storedString;

  function setString(string memory x) public {
    storedString = x;
  }

  function getString() public view returns (string memory retString) {
    return storedString;
  }
}

