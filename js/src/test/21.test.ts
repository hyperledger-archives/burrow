import * as assert from 'assert';
import * as test from '../test';

const Test = test.Test();

describe('issue #21', function () {
  this.timeout(10 * 1000);
  let contract: any;

  before(Test.before(function (burrow) {
    const source = `
      pragma solidity >=0.0.0;
      contract c {
        function getBytes() public pure returns (byte[10] memory){
            byte[10] memory b;
            string memory s = "hello";
            bytes memory sb = bytes(s);

            uint k = 0;
            for (uint i = 0; i < sb.length; i++) b[k++] = sb[i];
            b[9] = 0xff;
            return b;
        }

        function deeper() public pure returns (byte[12][100] memory s, uint count) {
          count = 42;
          return (s, count);
        }
      }
    `

    const {abi, bytecode} = test.compile(source, 'c')
    return burrow.contracts.deploy(abi, bytecode).then((c) => {
      contract = c
    })
  }))

  after(Test.after())

  it('gets the static byte array decoded properly', Test.it(function () {
    return contract.getBytes()
      .then((bytes) => {
        assert.deepEqual(
          bytes,
          [['68', '65', '6C', '6C', '6F', '00', '00', '00', '00', 'FF']]
        )
      })
  }))

  it('returns multiple values correctly from a function',
    Test.it(function () {
      return contract.deeper()
        .then((values) => {
          assert.equal(Number(values[1]), 42)
        })
    })
  )
})
