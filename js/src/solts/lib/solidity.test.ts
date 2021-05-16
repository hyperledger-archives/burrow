import assert from 'assert';
import { ABI } from './abi';
import { getSize, sha3, nameFromABI, tokenizeString } from './solidity';

describe('abi helpers', function () {
  it('should compute a valid method id', async function () {
    assert.equal(sha3('baz(uint32,bool)').slice(0, 8), 'CDCD77C0');
  });

  it('should return the full function name with args', async function () {
    const abi: ABI.Func = {
      type: 'function',
      name: 'baz',
      stateMutability: 'pure',
      inputs: [
        {
          name: '1',
          type: 'uint32',
        },
        {
          name: '2',
          type: 'bool',
        },
      ],
    };
    assert.equal(nameFromABI(abi), 'baz(uint32,bool)');
  });

  it('should strip array size', () => {
    assert.equal(getSize('uint[3]'), 3);
  });

  it('should tokenize string', () => {
    assert.equal(tokenizeString('sol/Storage.sol:Storage'), 'sol_Storage_sol_Storage');
  });
});
