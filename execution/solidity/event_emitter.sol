pragma solidity ^0.4.16;

contract EventEmitter {
    // indexed puts it in topic
    event ManyTypes(
        bytes32 indexed direction,
        bool trueism,
        string german ,
        int64 indexed newDepth,
        int bignum,
        string indexed hash);

    function EmitOne() public {
        emit ManyTypes("Downsie!", true, "Donaudampfschifffahrtselektrizitätenhauptbetriebswerkbauunterbeamtengesellschaft", 102, 42, "hash");
    }
}