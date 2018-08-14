pragma solidity ^0.4.20;

contract Concatenator {
    string _name;

    function getName() public view returns (string) {
        return _name;
    }

    function setName(string name) public {
        _name = name;
    }

    function addName(string name) public {
        _name = string(abi.encodePacked(_name, " ", name));
    }
}
