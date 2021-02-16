pragma solidity >0.6.3;

contract Child {

    address public owner; // public, so you can see it when you find the child

    constructor() public {
        owner = msg.sender;
    }
}
