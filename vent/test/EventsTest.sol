pragma solidity ^0.4.20;

contract EventsTest {

    event UpdateTestEvents(
        bytes32 indexed name,
        bytes32 indexed key,
        bytes32 indexed description);

    event DeleteTestEvents(
        bytes32 indexed name,
        bytes32 indexed key,
        int __DELETE__);

    bytes32 constant TABLE_EVENTS_TEST = "TEST_EVENTS";

    struct Thing {
        string name;
        string description;
        bool exists;
    }

    int length;
    mapping(string => Thing) things;

    function addThing(string _name, string _description) external {
        Thing storage thing = things[_name];
        if (!thing.exists) {
            length++;
        }
        thing.name = _name;
        thing.description = _description;
        thing.exists = true;
        emit UpdateTestEvents(prefix32(_name), TABLE_EVENTS_TEST, prefix32(_description));
    }

    function removeThing(string _name) external {
        Thing storage thing = things[_name];
        if (thing.exists) {
            length--;
            delete things[_name];
            emit DeleteTestEvents(prefix32(_name), TABLE_EVENTS_TEST, 0);
        }
    }

    function count() external view returns (int size) {
        return length;
    }

    function description(string _name) external view returns (string _description) {
        return things[_name].description;
    }

    function prefix32(string _str) private pure returns (bytes32 str32) {
        assembly {
        // We load one word256 after the start address of _str so that we get the first 32 bytes. Note: strings
        // with length less than 32 bytes are padded to a multiple of 32 bytes so we are not eating consecutive
        // memory here.
            str32 := mload(add(_str, 32))
        }
        return;
    }
}