pragma solidity >=0.0.0;

contract Zero {

  function zeroInt() public pure returns (int /* zeroInt */) {
    return 0;
  }

  function zeroUInt() public pure returns (uint /* zeroUInt */) {
    return 0;
  }

  function zeroBytes() public pure returns (bytes32 /* zeroBytes */) {
    return "";
  }

  function zeroAddress() public pure returns (address /* zeroAddress */) {
    return 0x0;
  }

  function zeroBool() public pure returns (bool /* zeroBool */) {
    return false;
  }
}
