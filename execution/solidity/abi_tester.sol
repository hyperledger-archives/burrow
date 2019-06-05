pragma solidity ^0.5.4;

contract A {
    function createB() public returns (B) {
        return new B();
    }
}

contract B {
    function createC() public returns (C) {
        return new C();
    }
}

contract C {
    uint public this_is_c;
}