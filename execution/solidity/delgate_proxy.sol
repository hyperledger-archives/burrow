pragma solidity ^0.5;

contract DelegateProxy {
    address internal proxied;

    function setDelegate(address _proxied) public {
        proxied = _proxied;
    }

    function getDelegate() public view returns (address) {
        return proxied;
    }
}
