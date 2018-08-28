pragma solidity >=0.0.0;

import "./base.sol";

contract Storage is base {
	constructor(uint x) base(x) public {}

	function test() public pure returns (uint) {
		return 42;
	}
}

