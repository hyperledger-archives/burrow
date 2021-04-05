import * as assert from 'assert';
import {burrow} from "../test";


describe('Namereg', function () {
  this.timeout(10 * 1000)

  it('Sets and gets a name correctly', async () => {
    await burrow.namereg.set('DOUG', 'ABCDEF0123456789', 5000, 100)
    const entry = await burrow.namereg.get('DOUG')
    assert.strictEqual(entry.getData(), 'ABCDEF0123456789')
  });
})
