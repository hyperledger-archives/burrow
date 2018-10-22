# Error 5: Memory out of bounds  
This fixture deploys and tests a proxy forwarding pattern implemented using assembly in the DefaultUserAccount contract. The included test performs mulitple forwardCall invocations and fails with a "Memory out of bounds" for functions that return a string.

The call opcode requires a valid argument in the 6th position; else we get an
memory out of bounds error.
