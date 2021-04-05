# WASM Contracts

Burrow supports experimental [ewasm](https://github.com/ewasm/design) contracts.
Any contract which can be compiled using [Solang](https://github.com/hyperledger-labs/solang)
can run on Burrow.

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