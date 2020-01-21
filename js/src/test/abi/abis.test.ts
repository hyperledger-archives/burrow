import * as assert from 'assert';
import * as test from '../../test';

const Test = test.Test();

describe('Abi', function () {
  this.timeout(10 * 1000)
  let burrow

  before(Test.before(function (_burrow) {
    burrow = _burrow
  }))

  after(Test.after())

  it('Call contract via burrow side Abi', Test.it(function () {
    return burrow.namereg.get('random')
      .then((data) => {
        let address = data.Data
        return burrow.contracts.address(address)
      })
      .then((contract) => {
        return contract.getRandomNumber()
      })
      .then((data) => {
        assert.equal(data[0], 55)
      })
  }))
})
