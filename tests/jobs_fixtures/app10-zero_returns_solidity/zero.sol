pragma solidity >=0.0.0;

contract Zero {

  function zeroInt() returns (int zeroInt) {
    return 0;
  }

  function zeroUInt() returns (uint zeroUInt) {
    return 0;
  }

  function zeroBytes() returns (bytes32 zeroBytes) {
    return "";
  }

  function zeroAddress() returns (address zeroAddress) {
    return 0x0;
  }

  function zeroBool() returns (bool zeroBool) {
    return false;
  }
}
