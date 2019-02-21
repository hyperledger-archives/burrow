pragma solidity ^0.5.4;
import * as foo from "contractImport4.sol";

contract map {
	mapping(uint=>uint) someMapping;
	function getMappingElement(uint a) public returns (uint) {
		return someMapping[a];
	}
}