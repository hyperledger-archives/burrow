import * as assert from 'assert';
import * as test from '../test';

const Test = test.Test();

describe('#81', function () {
  before(Test.before())
  after(Test.after())

  this.timeout(10 * 1000)

  it('listens to an event from a contract', Test.it(function (burrow) {
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

    const {abi, bytecode} = test.compile(source, 'Contract')
    return burrow.contracts.deploy(abi, bytecode)
      .then((contract: any) => {
        return new Promise((resolve, reject) => {
          const stream = contract.Pay((error, result) => {
            if (error) {
              reject(error)
            } else {
              try {
                const actual = Object.assign(
                  {},
                  result.args,
                  {amount: Number(result.args.amount)}
                )

                assert.deepEqual(
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
              } catch (exception) {
                reject(exception)
              }

              stream.cancel()
              resolve()
            }
          })

          contract.announce()
        })
      })
  }))
})
