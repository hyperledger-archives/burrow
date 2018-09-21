pragma solidity ^0.4.16;

contract Revert {
    function RevertIf0(uint32 i) public pure
    {
        if (i == 0) {
            revert("arbeidsongeschiktheidsverzekeringsmaatschappij");
        }
    }
}
