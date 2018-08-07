pragma solidity ^0.4.16;

contract Revert {
    function RevertAt(uint32 i) public {
        if (i == 0) {
            revert("I have reverted");
        } else {
            i--;
            this.RevertAt(i);
        }
    }
}