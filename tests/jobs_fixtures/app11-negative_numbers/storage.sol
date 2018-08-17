pragma solidity >=0.0.0;

contract testNegatives {
	int public x;
	int[2] negativeArray;
	function testNegatives(int a){
		x = a;
	} 
	function setX(int a){
		x = a;
	}
	function setNegativeArray(int a, int b) {
		negativeArray[0] = a;
		negativeArray[1] = b;
	}
	function getNegativeArray() returns (int[2]) {
		return negativeArray;
	}
}