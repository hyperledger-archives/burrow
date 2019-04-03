# Deploy Contracts

Now that the [burrow node is running](single-full-node.md), we can deploy contracts.

For this step, we need two things: one or more solidity contracts, and an `deploy.yaml`.

Let's take a simple example, found in [this directory](../../tests/jobs_fixtures/app06-deploy_basic_contract_and_different_solc_types_packed_unpacked/).

The `deploy.yaml` should look like:

```yaml
jobs:

- name: deployStorageK
  deploy:
      contract: storage.sol

- name: setStorageBaseBool
  set:
      val: "true"

- name: setStorageBool
  call:
      destination: $deployStorageK
      function: setBool
      data:
        - $setStorageBaseBool

- name: queryStorageBool
  query-contract:
      destination: $deployStorageK
      function: getBool

- name: assertStorageBool
  assert:
      key: $queryStorageBool
      relation: eq
      val: $setStorageBaseBool

# tests string bools: #71
- name: setStorageBool2
  call:
      destination: $deployStorageK
      function: setBool2
      data:
        - true

- name: queryStorageBool2
  query-contract:
      destination: $deployStorageK
      function: getBool2

- name: assertStorageBool2
  assert:
      key: $queryStorageBool2
      relation: eq
      val: "true"

- name: setStorageBaseInt
  set:
      val: 50000

- name: setStorageInt
  call:
      destination: $deployStorageK
      function: setInt
      data:
        - $setStorageBaseInt

- name: queryStorageInt
  query-contract:
      destination: $deployStorageK
      function: getInt

- name: assertStorageInt
  assert:
      key: $queryStorageInt
      relation: eq
      val: $setStorageBaseInt

- name: setStorageBaseUint
  set:
      val: 9999999

- name: setStorageUint
  call:
      destination: $deployStorageK
      function: setUint
      data:
        - $setStorageBaseUint

- name: queryStorageUint
  query-contract:
      destination: $deployStorageK
      function: getUint

- name: assertStorageUint
  assert:
      key: $queryStorageUint
      relation: eq
      val: $setStorageBaseUint

- name: setStorageBaseAddress
  set:
      val: "1040E6521541DAB4E7EE57F21226DD17CE9F0FB7"

- name: setStorageAddress
  call:
      destination: $deployStorageK
      function: setAddress
      data:
        - $setStorageBaseAddress

- name: queryStorageAddress
  query-contract:
      destination: $deployStorageK
      function: getAddress

- name: assertStorageAddress
  assert:
      key: $queryStorageAddress
      relation: eq
      val: $setStorageBaseAddress

- name: setStorageBaseBytes
  set:
      val: marmatoshi

- name: setStorageBytes
  call:
      destination: $deployStorageK
      function: setBytes
      data:
        - $setStorageBaseBytes

- name: queryStorageBytes
  query-contract:
      destination: $deployStorageK
      function: getBytes

- name: assertStorageBytes
  assert:
      key: $queryStorageBytes
      relation: eq
      val: $setStorageBaseBytes

- name: setStorageBaseString
  set:
      val: nakaburrow

- name: setStorageString
  call:
      destination: $deployStorageK
      function: setString
      data:
        - $setStorageBaseString

- name: queryStorageString
  query-contract:
      destination: $deployStorageK
      function: getString

- name: assertStorageString
  assert:
      key: $queryStorageString
      relation: eq
      val: $setStorageBaseString

```

while our Solidity contract (`storage.sol`) looks like:

```solidity
pragma solidity >=0.0.0;

contract SimpleStorage {
  bool storedBool;
  bool storedBool2;
  int storedInt;
  uint storedUint;
  address storedAddress;
  bytes32 storedBytes;
  string storedString;

  function setBool(bool x) public {
    storedBool = x;
  }

  function getBool() view public returns (bool retBool) {
    return storedBool;
  }

  function setBool2(bool x) public {
    storedBool2 = x;
  }

  function getBool2() view public returns (bool retBool) {
    return storedBool2;
  }

  function setInt(int x) public {
    storedInt = x;
  }

  function getInt() view public returns (int retInt) {
    return storedInt;
  }

  function setUint(uint x) public {
    storedUint = x;
  }

  function getUint() view public returns (uint retUint) {
    return storedUint;
  }

  function setAddress(address x) public {
    storedAddress = x;
  }

  function getAddress() view public returns (address retAddress) {
    return storedAddress;
  }

  function setBytes(bytes32 x) public {
    storedBytes = x;
  }

  function getBytes() view public returns (bytes32 retBytes) {
    return storedBytes;
  }

  function setString(string memory x) public {
    storedString = x;
  }

  function getString() view public returns (string memory retString) {
    return storedString;
  }
}
```

Both files (`deploy.yaml` & `storage.sol`) should be in the same directory with **no other yaml** or sol files.

You have to install [solc binary](https://solidity.readthedocs.io/en/v0.4.21/installing-solidity.html) in order to compile Solidity code.

From inside that directory, we are ready to deploy.

```bash
burrow deploy --address F71831847564B7008AD30DD56336D9C42787CF63 deploy.yaml
```

where you should replace the `--address` field with the `ValidatorAddress` at the top of your `burrow.toml`.

That's it! You've successfully deployed (and tested) a Solidity contract to a Burrow node.

Note - that to redeploy the burrow chain later, you will need the same genesis-spec.json and burrow.toml files - so keep hold of them!
