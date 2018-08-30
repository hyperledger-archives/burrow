pragma solidity ^0.4.16;

contract EventEmitter {
    // indexed puts it in topic
    event ManyTypes(
        bytes32 indexed direction,
        bool trueism,
        string german ,
        int indexed newDepth,
        string indexed hash)
        anonymous;

    function EmitOne() public {
        emit ManyTypes("Downsie!", true, "Donaudampfschifffahrtselektrizitätenhauptbetriebswerkbauunterbeamtengesellschaft", 102, 'hash');
    }
}