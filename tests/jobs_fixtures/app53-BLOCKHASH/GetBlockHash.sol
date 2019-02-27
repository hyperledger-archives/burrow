pragma solidity >=0.0.0;

contract GetBlockHash {

  function getBlockHash(uint blockNumber) view public returns (bytes32) {
    return blockhash(blockNumber);
  }
}