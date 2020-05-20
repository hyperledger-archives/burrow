import * as assert from 'assert';
import {burrow, compile} from '../test';

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
    `

      const {abi, code} = compile(source, 'Contract')
      const contract: any = await burrow.contracts.deploy(abi, code)
      const stream = contract.Pay((error, result) => {
        if (error) {
          throw error;
        } else {
          const actual = Object.assign(
            {},
            result.args,
            {amount: Number(result.args.amount)}
          )

          assert.deepStrictEqual(
            actual,
            {
              originator: '88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F',
              beneficiary: '721584FA4F1B9F51950018073A8E5ECF47F2D3B8',
              amount: 1,

              servicename: 'Energy',

              nickname: 'wasmachine',

              providername: 'Eneco',

              randomBytes: '000000000000000000000000000000000000000000000000DEADFEEDBEEFFACE'
            }
          )

          stream.cancel()
        }
      });

      contract.announce()
    })
});

