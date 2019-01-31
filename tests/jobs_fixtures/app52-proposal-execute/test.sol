pragma solidity ^0.4.25;

contract bar {
	uint foobar;

	function addFoobar(uint n) public {
		foobar += n;
	}

	function getFoobar() public view returns (uint n) {
		n = foobar;
	}
}
