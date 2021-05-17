pragma solidity >=0.0.0;

contract Contract {
    event Init(
        bytes32 indexed eventId,
        bytes32 indexed intervalId,
        address eventAddress,
        string namespace,
        string name,
        address controller,
        uint threshold,
        string metadata
    );

    function announce() public {
        emit Init(bytes32("event1"),
            bytes32("interval2"),
            0x59C99d4EbF520619ee7F806f11d90a9cac02CE06,
            "dining",
            "breakfast",
            msg.sender,
            4,
            "bacon,beans,eggs,tomato"
        );
    }
}
