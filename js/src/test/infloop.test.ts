import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { client } from './test';

describe('Really Long Loop', function () {
  let instance: any;
  this.timeout(1000000);

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
    `;

    const contract = compile(source, 'main');
    instance = await contract.deployWith(client, {
      middleware: (callTx) => {
        // Normal gas for deploy (when address === '')
        if (callTx.getAddress()) {
          // Hardly any for call
          callTx.setGaslimit(11);
        }
        return callTx;
      },
    });
  });

  it('It catches a revert when gas runs out', async () => {
    await instance
      .test()
      .then((str: string) => {
        throw new Error('Did not catch revert error');
      })
      .catch((err: Error) => {
        assert.match(err.message, /Error 5: insufficient gas/);
      });
  });
});
