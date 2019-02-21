pragma solidity >=0.0.0;

contract testNegatives {
	int public x;
	int[2] negativeArray;
	constructor(int a) public {
		x = a;
	} 
	function setX(int a) public {
		x = a;
	}
	function setNegativeArray(int a, int b) public {
		negativeArray[0] = a;
		negativeArray[1] = b;
	}
	function getNegativeArray() public view returns (int[2] memory) {
		return negativeArray;
	}
}