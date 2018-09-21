pragma solidity >=0.0.0;

contract SimpleStorage {
  bytes storedString;

  function setString(bytes x) public {
    storedString = x;
  }

  function getString() public constant returns (bytes retString) {
    return storedString;
  }
}

