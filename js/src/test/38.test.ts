import * as assert from 'assert';
import * as test from '../test';
import { Contract } from '..';

const Test = test.Test();

describe('#38', function () {
  before(Test.before())
  after(Test.after())

  this.timeout(10 * 1000)

  it('#38', Test.it((burrow) => {
    const source = `
      pragma solidity >=0.0.0;
      contract Contract {
          event Event();

          function announce() public {
              emit Event();
          }
      }
    `
    const {abi, bytecode} = test.compile(source, 'Contract')
    return burrow.contracts.deploy(abi, bytecode).then((contract) => {
      const secondContract: any = new Contract(abi, null, contract.address, burrow)

      return new Promise((resolve, reject) => {
        const stream = secondContract.Event.at(contract.address, function (error, event) {
          if (error) {
            reject(error)
          } else {
            stream.cancel()
            resolve(event)
          }
        })

        secondContract.announce()
      })
    })
  }))
})
