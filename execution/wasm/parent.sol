pragma solidity >0.6.3;

import "./child.sol";

contract Parent {


    address payable owner;
    address[] public children; // public, list, get a child address at row #


    constructor () public payable{
        owner = msg.sender;
    }


    function createChild() public {
        Child child = new Child();
        children.push(address(child)); // you can use the getter to fetch child addresses
    }
    
    function getChild(uint256 index) public returns (address) {
        return children[index];
    }
    
    function close() public { 
        selfdestruct(owner); 
    }
    
    function txPrice() public returns (uint) {
        return tx.gasprice;
    }
    
    function blockDifficulty() public returns (uint256) {
        return block.difficulty;
    }
}
