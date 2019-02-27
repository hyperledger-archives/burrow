pragma solidity ^0.5.4;

import {importedContract as poop} from './contractImport2.sol';
import "./contractImport3.sol";

contract c {
	poop ic;
	
	int A;
	int B;
	constructor(address Addr, int a, int b) public {
		ic = poop(Addr);
		A = a;
		B = b;
	}
	function add () public {
		A = ic.add(A, B);
	}
}