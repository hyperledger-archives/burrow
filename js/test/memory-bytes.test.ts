import * as assert from 'assert';
import {burrow, compile} from "./test";


describe('issue #21', function () {
  let contract: any;

  before(async () => {
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

    const {abi, code} = compile(source, 'c')
    return burrow.contracts.deploy(abi, code).then((c) => {
      contract = c
    })
  })

  it('gets the static byte array decoded properly', async () => {
    return contract.getBytes()
      .then((bytes) => {
        assert.deepStrictEqual(
          bytes,
          [['68', '65', '6C', '6C', '6F', '00', '00', '00', '00', 'FF']]
        )
      })
  })

  it('returns multiple values correctly from a function', function () {
    return contract.deeper()
      .then((values) => {
        assert.strictEqual(Number(values[1]), 42)
      })
  })
})
