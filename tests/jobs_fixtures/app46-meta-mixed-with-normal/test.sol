pragma solidity ^0.5.4;

contract Concatenator {
    string _name;

    function getName() public view returns (string memory) {
        return _name;
    }

    function setName(string memory name) public {
        _name = name;
    }

    function addName(string memory name) public {
        _name = string(abi.encodePacked(_name, " ", name));
    }
}
