pragma solidity ^0.4.20;

contract EventsTest {

    event UpdateTestEvents(
        bytes32 indexed name,
        bytes32 indexed key,
        bytes32 indexed description);

    bytes32 constant TABLE_EVENTS_TEST = "TEST_EVENTS";

    mapping(string => string) events;
    string[] eventNames;

    function addEvent(string _name, string _description) external {
        events[_name] = _description;
        eventNames.push(_name);
        emit UpdateTestEvents(TABLE_EVENTS_TEST, prefix32(_name), prefix32(_description));
    }

    function getNumberOfEvents() external view returns (uint size) {
        return eventNames.length;
    }

    function getEventNameAtIndex(uint _index) external view returns (string name) {
        return eventNames[_index];
    }

    function getEventData(string _name) external view returns (string description) {
        return events[_name];
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