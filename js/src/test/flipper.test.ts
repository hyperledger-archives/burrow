import * as assert from 'assert';
import fs from 'fs';
import { Contract } from '../contracts/contract';
import { burrow } from './test';

describe('Wasm flipper:', function () {
  let TestContract: any;

  before(async () => {
    const abi: any[] = JSON.parse(fs.readFileSync('src/test/flipper.abi', 'utf-8'));
    const wasm: string = fs.readFileSync('src/test/flipper.wasm').toString('hex');
    TestContract = await new Contract({ abi, bytecode: wasm }).deploy(burrow, true);
  });

  it('Flip', async () => {
    let output = await TestContract.get();
    assert.strictEqual(output[0], true);
    await TestContract.flip();
    output = await TestContract.get();
    assert.strictEqual(output[0], false);
  });
});
