# WASM Contracts

Burrow supports experimental WASM contracts. Specifically, any contract which can be compiled
using [solang](https://github.com/hyperledger-labs/solang) can run on Burrow.

## How to use

Write a simple solidity contract which is supported by solang. For example:

```solidity
contract foobar {
    uint64 foo;

    function setFoo(uint64 n) public {
        foo = n;
    }

    function getFoo() public returns (uint64) {
        return foo;
    }
}
```

And a deploy yaml:

```yaml
jobs:

- name: deployFoobar
  deploy:
    contract: foobar.sol

- name: setFoo
  call:
    destination: $deployFoobar
    function: setFoo
    data: [ 102 ]

- name: getFoo
  call:
    destination: $deployFoobar
    function: getFoo
```

Now run this script using:

```
burrow deploy --wasm -a Participant_0 deploy.yaml
```

## Implementation details

When using WASM, the same account model is used. The only different is that a WASM virtual machine
is used. When a contract is deployed, the WASM code is stored as the account code (this is different
from the EVM model where the constructor "returns" the deployed code). The WASM file which is deployed
must have two exported functions:

```solidity
void constructor(int32*)
int32 function(int32*)
```

When the contract is deployed, burrow calls the constructor function with the abi encoded arguments
stored in WASM memory, pointed to by the single argument. The abi encoded arguments are prefixed with
the length.

On executing a function call, the exported function called `function` is called. This takes the abi
encoded arguments just like the constructor, and returns a pointer to the abi encoded return values.

From the WASM code we can access contract storage via the following externals:

```solidity
void set_storage32(int32 key, int32 *ptr, int32 len);
void get_storage32(int32 key, int32 *ptr, int32 len);
```
