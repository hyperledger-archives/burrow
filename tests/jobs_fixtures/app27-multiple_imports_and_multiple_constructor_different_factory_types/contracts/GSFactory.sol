pragma solidity >=0.0.0;

import "./GSSingle.sol";
import "./GSMulti.sol";

contract GSFactory {
    address lastCreatedSingle;
    address lastCreatedMulti;

    function createSingle(uint initialValueSingle) returns (address GSAddrSingle) {
        lastCreatedSingle = new GSSingle(initialValueSingle);
        return lastCreatedSingle;
    }

    function getLastSingle() returns (address GSAddrSingle) {
        return lastCreatedSingle;
    }

    function createMulti(uint initialValueFirst, uint initialValueSecond) returns (address GSAddrMulti) {
        lastCreatedMulti = new GSMulti(initialValueFirst, initialValueSecond);
        return lastCreatedMulti;
    }

    function getLastMulti() returns (address GSAddrMulti) {
        return lastCreatedMulti;
    }
}
