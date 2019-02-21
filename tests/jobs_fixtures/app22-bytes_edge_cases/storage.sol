pragma solidity >=0.0.0;

contract SimpleStorage {
  bytes storedString;

  function setString(bytes memory x) public {
    storedString = x;
  }

  function getString() public view returns (bytes memory retString) {
    return storedString;
  }
}

