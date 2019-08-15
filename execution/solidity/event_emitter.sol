pragma solidity ^0.5.4;

contract EventEmitter {
    // indexed puts it in topic
    event ManyTypes(
        bytes32 indexed direction,
        bool trueism,
        string german ,
        int64 indexed newDepth,
        int bignum,
        string indexed hash);

    event ManyTypes2(
        bytes32 indexed direction,
        bool trueism,
        string german ,
        int128 indexed newDepth,
        int8 bignum,
        string indexed hash);

    function EmitOne() public {
        emit ManyTypes("Downsie!", true, "Donaudampfschifffahrtselektrizitätenhauptbetriebswerkbauunterbeamtengesellschaft", 102, 42, "hash");
    }

    function EmitTwo() public {
        emit ManyTypes2("Downsie!", true, "Donaudampfschifffahrtselektrizitätenhauptbetriebswerkbauunterbeamtengesellschaft", 102, 42, "hash");
    }
}