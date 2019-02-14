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

    function addThing(string memory description, string memory url, string memory filehash, string memory filename) public returns (uint) {
        things.push(Thing(msg.sender, description, url, filehash, filename));
        return 10;
    }
}
