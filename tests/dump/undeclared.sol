pragma solidity ^0.4.25;

contract test {
	event Bar(string a, int b);

	constructor() public {
		emit Bar("constructor", 0);
	}

	function foo() public returns (int) {
		int a = b + 3;
		int b = a + 7;

		int c = a * b;

		emit Bar("result", c);

		return c;
	}
}
