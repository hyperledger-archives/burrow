import * as assert from 'assert';
import {burrow, compile} from '../test';
import * as grpc from '@grpc/grpc-js';

describe('REVERT non-constant', function () {
  let contract

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
    `

    const {abi, code} = compile(source, 'c')
    return burrow.contracts.deploy(abi, code).then((c) => {
      contract = c
    })
  })

  it('It catches a revert with the revert string',
    async () => {
      return contract.getString(1)
        .then((str) => {
          throw new Error('Did not catch revert error')
        }).catch((err: grpc.ServiceError) => {
          assert.strictEqual(err.code, grpc.status.ABORTED)
        })
    })
})
