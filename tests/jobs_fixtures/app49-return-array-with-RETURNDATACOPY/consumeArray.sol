pragma solidity ^0.5;

import 'returnArray.sol';

contract ConsumeArray {
    uint256 public lowestSingleDigitPrime;

    constructor (ReturnArray producer) public {
        uint256 x = producer.singleDigitPrimes()[0];
        lowestSingleDigitPrime = x;
    }
}
