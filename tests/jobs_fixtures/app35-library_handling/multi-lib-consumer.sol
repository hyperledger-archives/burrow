pragma solidity >=0.0.0;

import "./multi-lib.sol";

contract c {
	intStructs.intStruct myIntStruct;
	constructor() public {
		myIntStruct = intStructs.intStruct(1, 2);
	}

	function basicFunctionReturn() public constant returns (uint x, uint y) {
		x = basicMath.add(myIntStruct.x, myIntStruct.y);
		y = basicMath.subtract(myIntStruct.x, myIntStruct.y);
	}
}