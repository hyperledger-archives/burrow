# WASM Contracts

burrow supports experimental wasm contracts. Any contract which can be compiled
using [solang](https://github.com/hyperledger-labs/solang) can be run on burrow.

## How to use

Write a simple solidity contract which is supported by solang. For example:

```
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

```
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

When using wasm, the same account model is used. The only different is that a wasm virtual machine
is used. When a contract is deployed, the wasm code is stored as the account code (this is different
from the EVM model where the constructor "returns" the deployed code). The wasm file which is deployed
must have two exported functions:

```
void constructor(int32*)
int32 function(int32*)
```
When the contract is deployed, burrow calls the constructor function with the abi encoded arguments
stored in wasm memory, pointed to by the single argument. The abi encoded arguments are prefixed with
the length.

On executing a function call, the exported function called `function` is called . This takes the abi
encoded arguments just like the constructor, and returns a pointer to the abi encoded return values.

From the wasm code we can access contract storage via the following externals:

```
void set_storage32(int32 key, int32 *ptr, int32 len);
void get_storage32(int32 key, int32 *ptr, int32 len);
```
