pragma solidity ^0.5.1;
 
contract D {
   address x;
   constructor(address z) public payable {
       x = z;
   }
}
contract X {
   address z;
   constructor(address y) public payable {
       D d = new D(y);
   }
}
contract C {
   function createD() public {
       X newX = new X(msg.sender);
   }
}