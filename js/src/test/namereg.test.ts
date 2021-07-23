import * as assert from 'assert';
import { client } from './test';

describe('Namereg', function () {
  this.timeout(10 * 1000);

  it('Sets and gets a name correctly', async () => {
    await client.namereg.set('DOUG', 'ABCDEF0123456789', 5000, 100);
    const entry = await client.namereg.get('DOUG');
    assert.strictEqual(entry.getData(), 'ABCDEF0123456789');
  });
});
