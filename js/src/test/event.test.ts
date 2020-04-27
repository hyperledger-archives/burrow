import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { ContractEvent } from '../contracts/contract';
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
    `;
    const contract = compile(source, 'Contract');
    const instance: any = await contract.deploy(burrow);
    let count = 0;

    const event = instance.Event as ContractEvent;
    const stream = event((error, event) => {
      if (error) {
        throw error;
      } else {
        assert.strictEqual(event?.args?.from?.length, 40);

        count++;

        if (count === 2) {
          stream.cancel();
        }
      }
    });

    instance.announce();
    instance.announce();
  });
});
