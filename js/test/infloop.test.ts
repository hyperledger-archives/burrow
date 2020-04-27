import * as assert from 'assert';
import { compile } from '../src/contracts/compile'
import { burrow } from './test';

describe.skip('Really Long Loop', function () {
  let contract
  this.timeout(1000000)

  before(async () => {
    const source = `
      pragma solidity >=0.0.0;
      contract main {
        function test() public returns (string memory) {
            c sub = new c();
            return sub.getString();
          }
        }
        contract c {
        string s = "secret";
        uint n = 0;
        function getString() public returns (string memory){
          for (uint i = 0; i < 10000000000000; i++) {
            n += 1;
          }
          return s;
        }
      }
    `

    const {abi, code} = compile(source, 'main')
    const c = await burrow.contracts.deploy(abi, code.bytecode);
    contract = c;
  })

  it('It catches a revert when gas runs out', async () => {
    return contract.test()
      .then((str) => {
        throw new Error('Did not catch revert error')
      }).catch((err) => {
        assert.strictEqual(err.message, 'ERR_EXECUTION_REVERT')
      })
  })
})
