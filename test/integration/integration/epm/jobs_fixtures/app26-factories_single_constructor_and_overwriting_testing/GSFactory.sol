pragma solidity >=0.0.0;

import "./GSContract.sol";

contract GSFactory {
    address lastCreated;
    function create(uint initialValue) returns (address GSAddr) {
        lastCreated = new GSContract(initialValue);
        return lastCreated;
    }

    function getLast() returns (address GSAddr) {
        return lastCreated;
    }
}
