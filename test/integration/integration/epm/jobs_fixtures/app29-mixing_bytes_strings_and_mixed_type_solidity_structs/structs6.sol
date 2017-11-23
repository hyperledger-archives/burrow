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

    function addThing(string description, string url, string filehash, bytes32 filename) returns (uint) {
        things.push(Thing(msg.sender, description, url, filehash, filename));
        return 10;
    }
}