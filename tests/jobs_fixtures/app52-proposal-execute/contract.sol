pragma solidity ^0.5;

contract foo {
	uint foobar;

	function addFoobar(uint n) public {
		foobar += n;
	}

	function getFoobar() public view returns (uint n) {
		n = foobar;
	}
}
