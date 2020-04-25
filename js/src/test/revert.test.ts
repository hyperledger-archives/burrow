import * as assert from 'assert';
import {burrow, compile} from '../test';
import * as grpc from '@grpc/grpc-js';

describe('REVERT constant', function () {
  let contract: any;

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
    `

    const {abi, code} = compile(source, 'c')
    contract = await burrow.contracts.deploy(abi, code)
  })


  it('gets the string when revert not called', async () => {
    return contract.getString(42)
      .then((str) => {
        assert.deepStrictEqual(str, ['secret'])
      })
  })

  it('It catches a revert with the revert string',
    async () => {
      return contract.getString(1)
        .then((str) => {
          throw new Error('Did not catch revert error')
        }).catch((err: grpc.ServiceError) => {
          assert.strictEqual(err.code, grpc.status.ABORTED)
          assert.strictEqual(err.message, '10 ABORTED: Did not pass correct key')
        })
    })
})
