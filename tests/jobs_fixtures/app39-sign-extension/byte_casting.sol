pragma solidity ^0.4.4;

contract ByteCasting {


      function Test(int8 _in1, int256 _in2, int16 _in3) public returns(int8 _out1, int256 _out2, int16 _out3) {

        bytes memory _buff = new  bytes(128);


        // Serializing
        uint128 _offst = 128;

        int8ToBytes(_offst,_in1,_buff);
        _offst -= 1;

        int256ToBytes(_offst,_in2,_buff);
        _offst -= 32;

        int16ToBytes(_offst,_in3,_buff);
        _offst -= 2;

        // Deserializing
        _offst = 128;

        _out1 = bytesToInt8(_offst,_buff);
        _offst -= 1;

        _out2 = bytesToInt256(_offst,_buff);
        _offst -= 32;

        _out3 = bytesToInt16(_offst,_buff);
        _offst -= 2;

    }


    function int8ToBytes(uint128 _offst, int8 _input, bytes _output) public {

        assembly {
            mstore(add(_output, _offst), _input)
        }
    }

    function int16ToBytes(uint128 _offst, int16 _input, bytes _output) public {

        assembly {
            mstore(add(_output, _offst), _input)
        }
    }

    function int256ToBytes(uint128 _offst, int256 _input, bytes _output) public {

        assembly {
            mstore(add(_output, _offst), _input)
        }
    }


    function bytesToInt8(uint128 _offst, bytes _input) public returns (int8 _output) {

        assembly {
            _output := mload(add(_input, _offst))
        }
    }

    function bytesToInt16(uint128 _offst, bytes _input) public returns (int16 _output) {

        assembly {
            _output := mload(add(_input, _offst))
        }
    }

    function bytesToInt256(uint128 _offst, bytes _input) public returns (int256 _output) {

        assembly {
            _output := mload(add(_input, _offst))
        }
    }
}