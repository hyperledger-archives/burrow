pragma solidity >=0.0.0;

contract Proxy {
    string name = '';

    constructor(string memory _name) public {
        name = _name;
    }

    function get() external view returns (string memory) {
        return name;
    }
}

contract Creator {
    function create(string calldata _name) external returns (address proxy) {
        return address(new Proxy(_name));
    }
}
