pragma solidity >=0.0.0;

contract Structs0 {
    struct Thing {
        address thingMaker;
        bytes32 description;
        bytes32 url;
        bytes32 filehash;
        bytes32 filename;
    }

    Thing[] things;

    function addThing(bytes32 description, bytes32 url, bytes32 filehash, bytes32 filename) returns (uint) {
        things.push(Thing(msg.sender, description, url, filehash, filename));
        return 10;
    }

    function getDesc() returns(bytes32 description) {
        return things[0].description;
    }

    function getUrl() returns(bytes32 url) {
        return things[0].url;
    }

    function getHash() returns(bytes32 filehash) {
        return things[0].filehash;
    }

    function getName() returns(bytes32 filename) {
        return things[0].filename;
    }
}
