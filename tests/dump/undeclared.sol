pragma solidity ^0.5.4;

contract test {
	event Bar(bytes32 a, int b);

	constructor() public {
		emit Bar("constructor", 0);
	}

	int foobar;

	function setFoobar(int n) public {
		foobar = n;
	}

	function getFoobar() view public {
		foobar;
	}

	function foo() public returns (int) {
		int a = 3;
		int b = a + 7;

		int c = a * b;

		emit Bar(hex"DEADCAFE", c);

		return c;
	}
}
