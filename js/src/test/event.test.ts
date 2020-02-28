import * as assert from 'assert';
import * as test from '../test';

const Test = test.Test();

describe('event', function () {
  before(Test.before())
  after(Test.after())

  this.timeout(10 * 1000)

  it('listens to an event from a contract', Test.it(function (burrow) {
    const source = `
      pragma solidity >=0.0.0;
      contract Contract {
          event Event(
              address from
          );

          function announce() public {
              emit Event(msg.sender);
          }
      }
    `
    const {abi, bytecode} = test.compile(source, 'Contract')
    return burrow.contracts.deploy(abi, bytecode)
      .then((contract: any) => {
        let count = 0;

        return new Promise((resolve, reject) => {
          const stream = contract.Event((error, event) => {
              if (error) {
                reject(error)
              } else {
                try {
                  assert.equal(event.args.from.length, 40)
                } catch (exception) {
                  reject(exception)
                }

                count++

                if (count === 2) {
                  stream.cancel()
                  resolve()
                }
              }
            })

          contract.announce()
          contract.announce()
        })
      })
  }))
})
