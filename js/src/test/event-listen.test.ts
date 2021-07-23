import { LogDescription } from '@ethersproject/abi';
import * as assert from 'assert';
import { compile } from '../contracts/compile';
import { ContractEvent } from '../contracts/contract';
import { client } from './test';

describe('Event listening', function () {
  it('listens to an event from a contract', async () => {
    const source = `
      pragma solidity >=0.0.0;
      contract Contract {
        event Pay(
          address originator,
          address beneficiary,
          int amount,
          string servicename,
          string nickname,
          string providername,
          bytes32 randomBytes
        );

        function announce() public {
          emit Pay(
            0x88977a37D05a4FE86d09E88c88a49C2fCF7d6d8F,
            0x721584fa4f1B9f51950018073A8E5ECF47f2d3b8,
            1,
            "Energy",
            "wasmachine",
            "Eneco",
            bytes32(uint256(0xDEADFEEDBEEFFACE))
          );
        }
      }
    `;

    const contract = compile(source, 'Contract');
    const instance = await contract.deploy(client);
    const promise = new Promise<LogDescription>((resolve, reject) => {
      const pay = instance.Pay as ContractEvent;
      const stream = pay((error, result) => {
        if (error || !result) {
          reject(error);
        } else {
          resolve(result);
          stream.cancel();
        }
      });
    });
    await instance.announce();

    const result = await promise;

    assert.strictEqual(result.args.originator, '88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F');
    assert.strictEqual(result.args.beneficiary, '721584FA4F1B9F51950018073A8E5ECF47F2D3B8');
    assert.strictEqual(Number(result.args.amount), 1);
    assert.strictEqual(result.args.servicename, 'Energy');
    assert.strictEqual(result.args.nickname, 'wasmachine');
    assert.strictEqual(result.args.providername, 'Eneco');
    const randomBytes = result.args.randomBytes as Buffer;
    assert.strictEqual(
      randomBytes.toString('hex').toUpperCase(),
      '000000000000000000000000000000000000000000000000DEADFEEDBEEFFACE',
    );
  });
});
