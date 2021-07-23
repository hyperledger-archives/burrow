pragma solidity >=0.0.0;

library NegationLib{
    function negate(int a) public pure returns (int inverse) {
        inverse = -a;
    }
}
