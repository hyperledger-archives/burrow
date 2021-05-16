pragma solidity >=0.0.0;

contract Contract {
    event Event (
        address from
    );

    function announce() public {
        emit Event(msg.sender);
    }
}