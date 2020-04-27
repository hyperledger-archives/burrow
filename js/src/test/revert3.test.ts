import * as grpc from '@grpc/grpc-js';
import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { burrow } from './test';

describe('REVERT non-constant', function () {
  let instance: any;

  before(async () => {
    const source = `
      pragma solidity >=0.0.0;
      contract c {
        string s = "secret";
        uint n = 0;
        function getString(uint key) public returns (string memory){
          if (key != 42){
            revert();
          } else {
            n = n + 1;
            return s;
          }
        }
      }
    `;

    instance = await compile(source, 'c').deploy(burrow);
  });

  it('It catches a revert with the revert string', async () => {
    return instance
      .getString(1)
      .then((str: any) => {
        throw new Error('Did not catch revert error');
      })
      .catch((err: grpc.ServiceError) => {
        assert.strictEqual(err.code, grpc.status.ABORTED);
      });
  });
});
