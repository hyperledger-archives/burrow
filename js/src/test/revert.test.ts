import * as grpc from '@grpc/grpc-js';
import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { client } from './test';

describe('REVERT constant', function () {
  let instance: any;

  before(async () => {
    const source = `
      pragma solidity >=0.0.0;
      contract c {
        string s = "secret";
        function getString(uint key) public view returns (string memory){
          if (key != 42){
            revert("Did not pass correct key");
          } else {
            return s;
          }
        }
      }
    `;

    instance = await compile(source, 'c').deploy(client);
  });

  it('gets the string when revert not called', async () => {
    return instance.getString(42).then((str: string) => {
      assert.deepStrictEqual(str, ['secret']);
    });
  });

  it('It catches a revert with the revert string', async () => {
    return instance
      .getString(1)
      .then((str: string) => {
        throw new Error('Did not catch revert error');
      })
      .catch((err: grpc.ServiceError) => {
        assert.strictEqual(err.code, grpc.status.ABORTED);
        assert.strictEqual(err.message, '10 ABORTED: Did not pass correct key');
      });
  });
});
