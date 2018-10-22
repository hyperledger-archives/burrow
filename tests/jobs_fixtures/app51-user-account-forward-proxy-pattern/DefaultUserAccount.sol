pragma solidity ^0.4.23;

/**
 * @title DefaultUserAccount
 * @dev The default implementation of a UserAccount
 */
contract DefaultUserAccount {

    /**
     * @dev Forwards a call to the specified target using the given bytes message.
     * @param _target the address to call
     * @param _payload the function payload consisting of the 4-bytes function hash and the abi-encoded function parameters which is typically created by
     * calling abi.encodeWithSelector(bytes4, args...) or abi.encodeWithSignature(signatureString, args...) 
     * @return success - whether the forwarding call returned normally
     * @return returnData - the bytes returned from calling the target function, if successful (NOTE: this is currently not supported, yet, and the returnData will always be empty)
     * REVERTS if:
     * - the target address is empty (0x0)
     */
    function forwardCall(address _target, bytes _payload)
        external
        returns (bool success, bytes returnData)
    {
        require(_target != address(0), "Empty target not allowed");
        bytes memory data = _payload;
        assembly {
            success := call(gas, _target, 0, add(data, 0x20), mload(data), 0, 0)
        }
        if (success) {
            uint returnSize;
            assembly {
                returnSize := returndatasize
            }
            returnData = new bytes(returnSize); // allocates a new byte array with the right size
            assembly {
                returndatacopy(add(returnData, 0x20), 0, returnSize) // copies the returned bytes from the function call into the return variable
            }
        }
    }

}
