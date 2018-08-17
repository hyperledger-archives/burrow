pragma solidity >=0.0.0;

library Search {
    function indexOf(uint[] storage self, uint value) returns (uint) {
        for (uint i = 0; i < self.length; i++)
            if (self[i] == value) return i;
        return uint(-1);
    }
}