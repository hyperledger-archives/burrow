pragma solidity >=0.0.0;

import "./GSContract.sol";

contract GSFactory {
    address lastCreated;
    function create(uint initialValue) public returns (address GSAddr) {
        lastCreated = address(new GSContract(initialValue));
        return lastCreated;
    }

    function getLast() public view returns (address GSAddr) {
        return lastCreated;
    }
}
