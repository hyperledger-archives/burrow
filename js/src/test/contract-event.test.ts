import { compile } from '../contracts/compile';
import { ContractEvent, getAddress } from '../contracts/contract';
import { burrow } from './test';

describe('Nested contract event emission', function () {
  it('#38', async () => {
    const source = `
      pragma solidity >=0.0.0;
      contract Contract {
          event Event();

          function announce() public {
              emit Event();
          }
      }
    `;
    const contract = compile(source, 'Contract');
    const instance: any = await contract.deploy(burrow);
    const event = instance.Event as ContractEvent;
    const stream = event.at(getAddress(instance), function (error, result) {
      if (error) {
        throw error;
      } else {
        stream.cancel();
      }
      instance.announce();
    });
  });
});
