pragma solidity >=0.0.0;

contract Contains {
    function contains(address[] memory _list, address _value) public pure returns (bool) {
	    for (uint i = 0; i < _list.length; i++) {
	        if (_list[i] == _value) return true;
	    }
	    return false;
    }

    function contains(uint[] memory _list, uint _value) public pure returns (bool) {
	    for (uint i = 0; i < _list.length; i++) {
	        if (_list[i] == _value) return true;
	    }
	    return false;
    }
}