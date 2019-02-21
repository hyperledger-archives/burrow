pragma solidity ^0.5.4;

contract test {
	event Bar(string a, int b);

	constructor() public {
		emit Bar("constructor", 0);
	}

	function foo() public returns (int) {
		int a = 3;
		int b = a + 7;

		int c = a * b;

		emit Bar("result", c);

		return c;
	}
}
