pragma solidity ^0.4.25;
interface AdoptionInterface{
	    function adopt(uint petId) external returns (uint);
	    function getAdopters() external view returns (address[16] memory);
	    function adopters(uint) external view returns (address);
	  }
	  
contract TestAdoption {
    
 event TestEvent(bool indexed name, string message);
	  
 AdoptionInterface adoption;

constructor(AdoptionInterface _adoption) public {
    adoption = _adoption;
    
}
 // Testing the adopt() function
 function testGetAdopterAddressByPetId() public returns (address result) {
   adoption.adopt(expectedPetId);
   return adoption.adopters(expectedPetId);
 }
 
 // The id of the pet that will be used for testing
 uint expectedPetId = 8;

 //The expected owner of adopted pet is this contract
 address expectedAdopter = address(this);

}

