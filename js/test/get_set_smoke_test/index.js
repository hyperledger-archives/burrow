const path = require('path')
const assert = require('assert')
const fs = require('fs-extra')

const test = require('../../lib/test')

const Test = test.Test()

const sourcePath = './GetSet.sol'

const source = fs.readFileSync(path.join(__dirname, sourcePath)).toString()
const {abi, bytecode} = test.compile(source, ':GetSet')

const testUint = 42
const testBytes = 'DEADBEEF00000000000000000000000000000000000000000000000000000000'
const testString = 'Hello World!'
const testBool = true

let TestContract

// Create a factory for the contract with the JSON interface 'myAbi'.

describe('Setting and Getting Values:', function () {
  this.timeout(10 * 1000)

  before(Test.before((burrow) => {
    return burrow.contracts.deploy(abi, bytecode).then((contract) => {
      TestContract = contract
    })
  }))
  after(Test.after())

  it('Uint', Test.it(function (burrow) {
    return new Promise((resolve, reject) => {
      TestContract.setUint(testUint, function (err) {
        if (err) { reject(err) }

        TestContract.getUint(function (err, output) {
          if (err) { reject(err) }

          assert.equal(output[0], testUint)
          resolve()
        })
      })
    })
  }))

  it('Bool', Test.it(function (burrow) {
    return new Promise((resolve, reject) => {
      TestContract.setBool(testBool, function (err) {
        if (err) { reject(err) }

        TestContract.getBool(function (err, output) {
          if (err) { reject(err) }

          assert.equal(output[0], testBool)
          resolve()
        })
      })
    })
  }))

  it('Bytes', Test.it(function (burrow) {
    return new Promise((resolve, reject) => {
      TestContract.setBytes(testBytes, function (err) {
        if (err) { reject(err) }

        TestContract.getBytes(function (err, output) {
          if (err) { reject(err) }

          assert.equal(output[0], testBytes)
          resolve()
        })
      })
    })
  }))

  it('String', Test.it(function (burrow) {
    return new Promise((resolve, reject) => {
      TestContract.setString(testString, function (err) {
        if (err) { reject(err) }

        TestContract.getString(function (err, output) {
          if (err) { reject(err) }

          assert.equal(output[0], testString)
          resolve()
        })
      })
    })
  }))
})
