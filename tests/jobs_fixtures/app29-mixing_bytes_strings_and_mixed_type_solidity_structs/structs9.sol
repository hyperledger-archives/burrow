pragma solidity >=0.0.0;

contract Structs9 {
    struct Thing {
        address thingMaker;
        bytes32 description;
        string url;
        bytes32 filehash;
        string filename;
    }

    Thing[] things;

    function addThing(bytes32 description, string url, bytes32 filehash, string filename) returns (uint) {
        things.push(Thing(msg.sender, description, url, filehash, filename));
        return 10;
    }
}