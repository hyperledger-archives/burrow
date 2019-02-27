pragma solidity ^0.5.4;

contract Revert {
    function RevertIf0(uint32 i) public pure
    {
        if (i == 0) {
            revert("arbeidsongeschiktheidsverzekeringsmaatschappij");
        }
    }
}
