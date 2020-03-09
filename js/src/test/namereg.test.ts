import * as assert from 'assert';
import * as test from '../test';
import { Burrow } from '..';

const Test = test.Test();

describe('Namereg', function () {
  this.timeout(10 * 1000)
  let burrow: Burrow;

  before(Test.before(function (_burrow) {
    burrow = _burrow
  }))

  after(Test.after())

  it('Sets and gets a name correctly', Test.it(function () {
    return burrow.namereg.set('DOUG', 'ABCDEF0123456789', 5000, 100, (err, exec) => {
      
      return burrow.namereg.get('DOUG', (err, exec) => {
        assert.equal(exec.getData(), 'ABCDEF0123456789')
      })
    })
  }));
})
