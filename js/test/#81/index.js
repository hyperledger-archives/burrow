'use strict'

const assert = require('assert')
const test = require('../../lib/test')

const Test = test.Test()

describe('#81', function () {
  before(Test.before())
  after(Test.after())

  this.timeout(10 * 1000)

  it('listens to an event from a contract', Test.it(function (burrow) {
    const source = `
      pragma solidity ^0.4.21;
      contract Contract {
        event Pay(
          address originator,
          address beneficiary,
          int amount,
          string servicename,
          string alias,
          string providername,
          bytes32 randomBytes
        );

        function emit() public {
          emit Pay(
            0x88977a37D05a4FE86d09E88c88a49C2fCF7d6d8F,
            0x721584fa4f1B9f51950018073A8E5ECF47f2d3b8,
            1,
            "Energy",
            "wasmachine",
            "Eneco",
            0xDEADFEEDBEEFFACE
          );
        }
      }
    `

    const {abi, bytecode} = test.compile(source, ':Contract')
    return burrow.contracts.deploy(abi, bytecode)
      .then((contract) => {
        return new Promise((resolve, reject) => {
          contract.Pay((error, {args}) => {
            if (error) {
              reject(error)
            } else {
              try {
                const actual = Object.assign(
                  {},
                  args,
                  {amount: Number(args.amount)}
                )

                assert.deepEqual(
                  actual,
                  {
                    originator: '88977A37D05A4FE86D09E88C88A49C2FCF7D6D8F',
                    beneficiary: '721584FA4F1B9F51950018073A8E5ECF47F2D3B8',
                    amount: 1,

                    servicename: 'Energy',

                    alias: 'wasmachine',

                    providername: 'Eneco',

                    randomBytes: '000000000000000000000000000000000000000000000000DEADFEEDBEEFFACE'
                  }
                )
              } catch (exception) {
                reject(exception)
              }

              resolve()
            }
          })

          contract.emit()
        })
      })
  }))
})
