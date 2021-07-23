import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { client } from './test';

describe('memory bytes', function () {
  let instance: any;

  before(async () => {
    const source = `
      pragma solidity >=0.0.0;
      contract c {
        function getBytes() public pure returns (bytes1[10] memory){
            bytes1[10] memory b;
            string memory s = "hello";
            bytes memory sb = bytes(s);

            uint k = 0;
            for (uint i = 0; i < sb.length; i++) b[k++] = sb[i];
            b[9] = 0xff;
            return b;
        }

        function deeper() public pure returns (bytes1[12][100] memory s, uint count) {
          count = 42;
          return (s, count);
        }
      }
    `;

    const contract = compile(source, 'c');
    instance = await contract.deploy(client);
  });

  it('gets the static byte array decoded properly', async () => {
    const [bytes] = await instance.getBytes();
    assert.deepStrictEqual(
      bytes.map((b: Buffer) => b.toString('hex').toUpperCase()),
      ['68', '65', '6C', '6C', '6F', '00', '00', '00', '00', 'FF'],
    );
  });

  it('returns multiple values correctly from a function', async () => {
    const values = await instance.deeper();
    assert.strictEqual(values[1], 42);
  });
});
