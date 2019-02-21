pragma solidity >=0.0.0;

import "./DefaultUserAccount.sol";

contract UserAccountTest {

	string constant SUCCESS = "success";
	string longString = "longString";

	string testServiceFunctionSig = "serviceInvocation(address,uint256,bytes32)";

	/**
	 * @dev Tests the DefaultUserAccount call forwarding logic
	 */
	function testCallForwarding() external returns (string memory) {

		uint testState = 42;
		bytes32 testKey = "myKey";
		TestService testService = new TestService();
		bool success;

		DefaultUserAccount account = new DefaultUserAccount();

		// The bytes payload encoding the function signature and parameters for the forwarding call
		bytes memory payload = abi.encodeWithSignature(testServiceFunctionSig, address(this), testState, testKey);

		// test failures
		// *IMPORTANT*: the use of the abi.encode function for this call is extremely important since sending the parameters individually via call(bytes4, args...)
		// has known problems encoding the dynamic-size parameters correctly, see https://github.com/ethereum/solidity/issues/2884
		(success, ) = address(account).call(abi.encodeWithSelector(bytes4(keccak256("forwardCall(address,bytes)")), address(0), payload));
		if (success)
			return "Forwarding a call to an empty address should revert";
		(success, ) = account.forwardCall(address(testService), abi.encodeWithSignature("fakeFunction(bytes32)", testState));
		if (success)
			return "Forwarding a call to a non-existent function should return false";

		// test successful invocation
		bytes memory returnData;
		(success, returnData) = account.forwardCall(address(testService), payload);
		if (!success) return "Forwarding a call from an authorized address with correct payload should return true";
		if (testService.currentEntity() != address(this)) return "The testService should show this address as the current entity";
		if (testService.currentState() != testState) return "The testService should have the testState set";
		if (testService.currentKey() != testKey) return "The testService should have the testKey set";
		if (testService.lastCaller() != address(account)) return "The testService should show the DefaultUserAccount as the last caller";
		if (returnData.length != 32) return "ReturnData should be of size 32";
		// TODO ability to decode return data via abi requires 0.5.0.
		// (bytes32 returnMessage) = abi.decode(returnData,(bytes32));
		if (toBytes32(returnData, 32) != testService.getSuccessMessage()) return "The function return data should match the service success message";

		// test different input/return data
		payload = abi.encodeWithSignature("isStringLonger5(string)", longString);
		(success, returnData) = account.forwardCall(address(testService), payload);
		if (!success) return "isStringLonger5 invocation should succeed";
		if (returnData[31] != hex"01") return "isStringLonger5 should return true for longString"; // boolean is left-padded, so the value is at the end of the bytes

		payload = abi.encodeWithSignature("getString()");
		(success, returnData) = account.forwardCall(address(testService), payload);
		if (!success) return "getString invocation should succeed";
		(string memory RetString) = abi.decode(returnData, (string));
		string memory expected = "Hello World";
		// Solidity thinks string compare should be as hard as possible.
		if (keccak256(abi.encodePacked(RetString)) != keccak256(abi.encodePacked(expected))) return "getString should return Hello World";

		return SUCCESS;
	}

    function toBytes32(bytes memory b, int offset) public pure returns (bytes32 result) {
	    assembly {
    	    result := mload(add(b, offset))
    	}
    }

}

/**
 * @dev Contract providing typical service functions to use as target for call forwarding.
 */
contract TestService {

	address public currentEntity;
	uint public currentState;
	bytes32 public currentKey;
	address public lastCaller;
	string public storedString = "Hello World";

	function serviceInvocation(address _entity, uint _newState, bytes32 _key) public returns (bytes32) {

		currentEntity = _entity;
		currentState = _newState;
		currentKey = _key;
		lastCaller = msg.sender;
		return "congrats";
	}

	function getString() public view returns (string memory) {
		return storedString;
	}

	function isStringLonger5(string memory _string) public pure returns (bool) {
		if (bytes(_string).length > 5)
			return true;
		else
			return false;
	} 

	function getSuccessMessage() public pure returns (bytes32) {
		return "congrats";	
	}
}
