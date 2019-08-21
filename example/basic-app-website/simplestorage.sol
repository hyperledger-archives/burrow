pragma solidity >=0.0.0;


contract simplestorage {
    uint public storedData;

    constructor(uint initVal) public {
        storedData = initVal;
    }

    function set(uint value) public {
        storedData = value;
    }

    function get() public view returns (uint value) {
        return storedData;
    }

    // Since transactions are executed atomically we can implement this concurrency primitive in Solidity with the
    // desired behaviour
    function testAndSet(uint expected, uint newValue) public returns (uint value, bool success) {
        if (storedData == expected) {
            storedData = newValue;
            return (storedData, true);
        }
        return (storedData, false);
    }
}
