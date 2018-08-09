pragma solidity ^0.4.0;

import {importedContract as poop} from './contractImport2.sol';
import "./contractImport3.sol";

contract c {
	poop ic;
	
	int A;
	int B;
	function c(address Addr, int a, int b) {
		ic = poop(Addr);
		A = a;
		B = b;
	}
	function add () {
		A = ic.add(A, B);
	}
}