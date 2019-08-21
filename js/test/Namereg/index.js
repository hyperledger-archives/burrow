'use strict'

const assert = require('assert')
const test = require('../../lib/test')

const Test = test.Test()

describe('Namereg', function () {
  this.timeout(10 * 1000)
  let burrow

  before(Test.before(function (_burrow) {
    burrow = _burrow
  }))

  after(Test.after())

  it('Sets and gets a name correctly', Test.it(function () {
    return burrow.namereg.set('DOUG', 'ABCDEF0123456789', 20)
      .then(() => {
        return burrow.namereg.get('DOUG')
          .then((data) => {
            assert.equal(data.Data, 'ABCDEF0123456789')
          })
      })
  }))
})
