pragma solidity >=0.0.0;

import "NegationLib.sol";

contract Addition {
    function add(int a, int b) public pure returns (int sum) {
        sum = a + b;
    }

    function sub(int a, int b) public pure returns (int) {
        return add(a, NegationLib.negate(b));
    }
}
