pragma solidity >=0.0.0;

contract Structs10 {
    struct Thing {
        address thingMaker;
        string description;
        bytes32 url;
        string filehash;
        bytes32 filename;
    }

    Thing[] things;

    function addThing(string description, bytes32 url, string filehash, bytes32 filename) returns (uint) {
        things.push(Thing(msg.sender, description, url, filehash, filename));
        return 10;
    }
}