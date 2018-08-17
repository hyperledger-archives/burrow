contract SimpleStorage {
  int storedData;

  function set(int x) {
    storedData = x;
  }

  function get() constant returns (int retVal) {
    return storedData;
  }
}