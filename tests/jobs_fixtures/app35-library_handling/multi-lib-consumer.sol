pragma solidity >=0.0.0;

import "./multi-lib.sol";

contract c {
	intStructs.intStruct myIntStruct;
	function c() {
		myIntStruct = intStructs.intStruct(1, 2);
	}

	function basicFunctionReturn() constant returns (uint x, uint y) {
		x = basicMath.add(myIntStruct.x, myIntStruct.y);
		y = basicMath.subtract(myIntStruct.x, myIntStruct.y);
	}
}