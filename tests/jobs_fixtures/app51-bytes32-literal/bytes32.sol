pragma solidity ^0.4.20;

contract bar {
  function getLiteralLength() public pure returns (int retInt) {
    return baz("");
  }

  function baz(bytes32 foo) public pure returns (int retInt) {
    if (foo == 0 && foo == "") {
      return 0;
    }

    return 102;
  }
}
