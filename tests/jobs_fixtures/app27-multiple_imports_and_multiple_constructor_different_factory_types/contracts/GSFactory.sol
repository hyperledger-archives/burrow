pragma solidity >=0.0.0;

import "./GSSingle.sol";
import "./GSMulti.sol";

contract GSFactory {
    address lastCreatedSingle;
    address lastCreatedMulti;

    function createSingle(uint initialValueSingle) public returns (address GSAddrSingle) {
        lastCreatedSingle = address(new GSSingle(initialValueSingle));
        return lastCreatedSingle;
    }

    function getLastSingle() public view returns (address GSAddrSingle) {
        return lastCreatedSingle;
    }

    function createMulti(uint initialValueFirst, uint initialValueSecond) public returns (address GSAddrMulti) {
        lastCreatedMulti = address(new GSMulti(initialValueFirst, initialValueSecond));
        return lastCreatedMulti;
    }

    function getLastMulti() public view returns (address GSAddrMulti) {
        return lastCreatedMulti;
    }
}
