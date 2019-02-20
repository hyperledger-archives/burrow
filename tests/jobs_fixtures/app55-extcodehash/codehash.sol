pragma solidity >=0.0.0;

contract foo {
	string str;
	constructor(string memory bar) public {
		bar = str;
    	}
}

contract bar {
    function bar2() public {
        foo f1 = new foo("abc");
        foo f2 = new foo("def");
	address a1 = address(f1);
	address a2 = address(f2);
        uint hash1;
        uint hash2;
        
        assembly {
            hash1 := extcodehash(a1)
            hash2 := extcodehash(a2)
        }
        assert(hash1 == hash2);
    }
}
