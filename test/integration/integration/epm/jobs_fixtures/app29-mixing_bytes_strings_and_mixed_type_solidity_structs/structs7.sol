pragma solidity >=0.0.0;

contract Structs7 {
    struct Thing {
        address thingMaker;
        bytes32 description;
        bytes32 url;
        string filehash;
        string filename;
    }

    Thing[] things;

    function addThing(bytes32 description, bytes32 url, string filehash, string filename) returns (uint) {
        things.push(Thing(msg.sender, description, url, filehash, filename));
        return 10;
    }
}