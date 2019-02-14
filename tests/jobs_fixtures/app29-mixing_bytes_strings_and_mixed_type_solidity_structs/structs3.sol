pragma solidity >=0.0.0;

contract Structs3 {
    struct Thing {
        address thingMaker;
        bytes32 description;
        string url;
        string filehash;
        string filename;
    }

    Thing[] things;

    function addThing(bytes32 description, string memory url, string memory filehash, string memory filename) public returns (uint) {
        things.push(Thing(msg.sender, description, url, filehash, filename));
        return 10;
    }
}
