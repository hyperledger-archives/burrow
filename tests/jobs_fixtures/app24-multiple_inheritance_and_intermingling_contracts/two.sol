pragma solidity >=0.0.0;

import "./one.sol";

contract two is one{

  function ii() public pure returns (uint /* id */) {
    return 2;
  }

}