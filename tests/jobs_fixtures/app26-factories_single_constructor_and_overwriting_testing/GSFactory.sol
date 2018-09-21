pragma solidity >=0.0.0;

import "./GSContract.sol";

contract GSFactory {
    address lastCreated;
    function create(uint initialValue) public returns (address GSAddr) {
        lastCreated = new GSContract(initialValue);
        return lastCreated;
    }

    function getLast() public view returns (address GSAddr) {
        return lastCreated;
    }
}
