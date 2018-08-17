pragma solidity >=0.0.0;

library basicMath {
	function add(uint x, uint y) returns (uint z) {
		z = x + y;
	}

	function subtract(uint x, uint y) returns (uint z) {
		z = x - y;
	}
}

library intStructs {
	struct intStruct {
		uint x;
		uint y;
	}
}