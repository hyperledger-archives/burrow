interface E {
	function get_vm() external returns (string memory);
	function get_number() external returns (int);
}

contract evm is E {
	function get_vm() public returns (string memory) {
		return "evm";
	}

	function get_number() public returns (int) {
		return 102;
	}
	
	function call_get_vm(E e) public returns (string memory) {
		return string(abi.encodePacked("evm called ", e.get_vm()));
	}

	function call_get_number(E e) public returns (int) {
		return e.get_number();
	}
}
