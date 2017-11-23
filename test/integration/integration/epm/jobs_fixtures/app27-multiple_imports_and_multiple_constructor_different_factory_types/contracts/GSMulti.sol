pragma solidity >=0.0.0;

contract GSMulti {
  uint storedData1;
  uint storedData2;

  function GSMulti(uint initialValue1, uint initialValue2) {
    storedData1 = initialValue1;
    storedData2 = initialValue2;
  }

  function set(uint first, uint second) {
    storedData1 = first;
    storedData2 = second;
  }

  function getFirst() constant returns (uint retVal) {
    return storedData1;
  }

  function getSecond() constant returns (uint retVal) {
    return storedData2;
  }
}
