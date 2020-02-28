import * as assert from 'assert';
import * as test from '../test';
import * as grpc from 'grpc';

const Test = test.Test();

describe('REVERT non-constant', function () {
  this.timeout(10 * 1000)
  let contract: any;

  before(Test.before(function (burrow) {
    const source = `
      pragma solidity >=0.0.0;
      contract c {
        string s = "secret";
        uint n = 0;
        function getString(uint key) public returns (string memory){
          if (key != 42){
            revert("Did not pass correct key");
          } else {
            n = n + 1;
            return s;
          }
        }
      }
    `

    const {abi, bytecode} = test.compile(source, 'c')
    return burrow.contracts.deploy(abi, bytecode).then((c) => {
      contract = c
    })
  }))

  after(Test.after())

  it('gets the string when revert not called', Test.it(function () {
    return contract.getString(42)
      .then((str) => {
        assert.equal(str, 'secret')
      })
  }))

  it('It catches a revert with the revert string',
    Test.it(function () {
      return contract.getString(1)
        .then((str) => {
          throw new Error('Did not catch revert error')
        }).catch((err: grpc.ServiceError) => {
          assert.equal(err.code, grpc.status.ABORTED)
          assert.equal(err.message, 'Did not pass correct key')
        })
    })
  )
})
