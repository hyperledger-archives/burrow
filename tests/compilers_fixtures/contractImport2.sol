pragma solidity ^0.5;
import "contractImport3.sol" as moarPoop;

contract importedContract {
	function add(int a, int b) public returns (int) {
		return a + b;
	}
	function subtract(int a, int b) public returns (int) {
		return a - b;
	}
	function addFromMapping(uint a, uint b) public returns (uint) {
		moarPoop.map map;
		return map.getMappingElement(a) + b;
	}
}