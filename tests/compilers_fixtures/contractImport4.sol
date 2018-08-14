pragma solidity ^0.4.0;

contract bar {
	struct Mod {
		uint x;
		uint k;
		uint m;
	}

	Mod mul;

	function getVariables() returns (uint, uint) {
		mul = Mod(1, 2, 3);
		return (mul.x, mul.k);
	}
}