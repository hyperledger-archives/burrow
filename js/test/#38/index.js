'use strict'

const test = require('../../lib/test')

const Test = test.Test()

describe('#38', function () {
  before(Test.before())
  after(Test.after())

  this.timeout(10 * 1000)

  it('#38', Test.it(function (burrow) {
    const source = `
      pragma solidity ^0.4.21;
      contract Contract {
          event Event();

          function emit() public {
              emit Event();
          }
      }
    `
    const {abi, bytecode} = test.compile(source, ':Contract')
    return burrow.contracts.deploy(abi, bytecode).then((contract) => {
      const secondContract = burrow.contracts.new(abi, null, contract.address)

      return new Promise((resolve, reject) => {
        secondContract.Event.once(function (error, event) {
          if (error) {
            reject(error)
          } else {
            resolve(event)
          }
        })

        secondContract.emit()
      })
    })
  }))
})
