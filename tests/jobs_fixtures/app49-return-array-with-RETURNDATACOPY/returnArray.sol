pragma solidity ^0.4.23;

contract ReturnArray {
    function singleDigitPrimes() pure external returns (uint256[] memory) {
        uint256[] memory retval = new uint256[](4);
        retval[0] = 2;
        retval[1] = 3;
        retval[2] = 5;
        retval[3] = 7;
        return retval;
    }
}
