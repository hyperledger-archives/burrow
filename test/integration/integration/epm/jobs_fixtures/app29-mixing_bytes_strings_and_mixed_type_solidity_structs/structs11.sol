pragma solidity >=0.0.0;

contract Structs11 {
    struct Thing {
        address thingMaker;
        bytes32 description;
        string url;
        string filehash;
        bytes32 filename;
    }

    Thing[] things;

    function addThing(bytes32 description, string url, string filehash, bytes32 filename) returns (uint) {
        things.push(Thing(msg.sender, description, url, filehash, filename));
        return 10;
    }
}
