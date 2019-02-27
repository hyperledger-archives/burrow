pragma solidity >=0.0.0;

contract Structs6 {
    struct Thing {
        address thingMaker;
        string description;
        string url;
        string filehash;
        bytes32 filename;
    }

    Thing[] things;

    function addThing(string memory description, string memory url, string memory filehash, bytes32 filename) public returns (uint) {
        things.push(Thing(msg.sender, description, url, filehash, filename));
        return 10;
    }
}