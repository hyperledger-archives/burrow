pragma solidity ^0.5.4;

contract bar {
	struct Mod {
		uint x;
		uint k;
		uint m;
	}

	Mod mul;

	function getVariables() public returns (uint, uint) {
		mul = Mod(1, 2, 3);
		return (mul.x, mul.k);
	}
}