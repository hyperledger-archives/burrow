import * as assert from 'assert';
import { compile } from '../src/contracts/compile'
import { burrow } from './test';

describe('event', function () {
  it('listens to an event from a contract', async () => {
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
    const {abi, code} = compile(source, 'Contract')
    const contract: any = await burrow.contracts.deploy(abi, code)
    let count = 0;

    const stream = contract.Event((error, event) => {
      if (error) {
        throw(error)
      } else {
        assert.strictEqual(event.args.from.length, 40)

        count++

        if (count === 2) {
          stream.cancel()
        }
      }
    })

    contract.announce()
    contract.announce()
  })
})
