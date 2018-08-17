pragma solidity ^0.4.16;

contract Revert {
    event NotReverting(uint32 indexed i);

    function RevertAt(uint32 i) public {
        if (i == 0) {
            revert("I have reverted");
        } else {
            i--;
            emit NotReverting(i);
            this.RevertAt(i);
        }
    }
}