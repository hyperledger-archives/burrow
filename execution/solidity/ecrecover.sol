pragma solidity ^0.5;

contract ECRecover {

    function recoverSigningAddress(bytes32 msgDigest, uint8 v, bytes32 r, bytes32 s) pure public returns (address) {
        return ecrecover(msgDigest, v, r, s);
    }

}
