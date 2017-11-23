pragma solidity >=0.0.0;

contract Structs1 {
    struct Thing {
        address thingMaker;
        string description;
        string url;
        string filehash;
        string filename;
    }

    Thing[] things;

    function addThing(string description, string url, string filehash, string filename) returns (uint) {
        things.push(Thing(msg.sender, description, url, filehash, filename));
        return 10;
    }
}
