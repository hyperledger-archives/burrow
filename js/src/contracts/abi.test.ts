import * as assert from 'assert';
import { transformToFullName } from "./abi";
import { compile } from "./compile";

const source = `
pragma solidity >=0.0.0;

contract foo {
	uint foobar;

	function addFoobar(uint n, bool b) public {
	  if (b) {
		  foobar += n;
	  }
	}

	function getFoobar() public view returns (uint n) {
		n = foobar;
	}
}
`

describe('ABI helpers', () => {
  it('Can extract correct names from ABI', () => {
    const {abi} = compile(source, 'foo')
    console.log(JSON.stringify(abi))
    assert.strictEqual(transformToFullName(abi[0]), "addFoobar(uint256,bool)")
  })

  it('Derive Typescript contract type from ABI', () => {
    // This is a nearly-working proof of concept for generating a contract's type definition
    // directly from the JSON ABI...

    const abiConst = [
      {
        "name": "addFoobar",
        "inputs": [
          {"internalType": "uint256", "name": "n", "type": "uint256"},
          {"internalType": "bool", "name": "b", "type": "bool"}
        ],
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
      },
      {
        "name": "getFoobar",
        "inputs": [
          {"internalType": "address", "name": "addr", "type": "address"}
        ],
        "outputs": [{"internalType": "uint256", "name": "n", "type": "uint256"}],
        "stateMutability": "view",
        "type": "function"
      },
    ] as const;

    type abi = typeof abiConst

    type Function = abi[number]

    type FunctionName = Function['name']

    type Picker<name extends FunctionName> = {
      [i in keyof abi]: abi[i] extends { name: name; type: 'function' } ? abi[i] : never
    };

    type ValueOf<T> = T[keyof T]

    type PickFunctionByName<name extends FunctionName> = ValueOf<Picker<name>>

    type Address = string

    type Type<T> =
      T extends 'uint256' ? number :
        T extends 'bool' ? boolean :
          T extends 'address' ? Address :
            never

    // **mumble** something about distribution
    type PickValue<T, U> = U extends keyof T ? Pick<T, U>[keyof Pick<T, U>] : never;

    type TypesOf<T> = { [k in keyof T]: Type<PickValue<T[k], 'type'>> }

    type FunctionInputs<T extends FunctionName> = PickFunctionByName<T>["inputs"]

    type FunctionOutputs<T extends FunctionName> = PickFunctionByName<T>["outputs"]

    const getFoobarABI: PickFunctionByName<'getFoobar'> =
      {
        "name": "getFoobar",
        "inputs": [
          {"internalType": "address", "name": "addr", "type": "address"}
        ],
        "outputs": [{"internalType": "uint256", "name": "n", "type": "uint256"}],
        "stateMutability": "view",
        "type": "function"
      }

    const addFoobarInputs: FunctionInputs<'addFoobar'> = [
      {"internalType": "uint256", "name": "n", "type": "uint256"},
      {"internalType": "bool", "name": "b", "type": "bool"},
    ]

    const addFoobarArgs: TypesOf<FunctionInputs<'addFoobar'>> = [1, true]
    const getFoobarArgs: TypesOf<FunctionInputs<'getFoobar'>> = ["address"]


    // Everything above this line comiles, which make me think the following should work, however...

    type Contract = {
      // This line is a compiler error: TS2370: A rest parameter must be of an array type.
      // Uncomment to experiment further
      // [name in FunctionName]: (...args: TypesOf<FunctionInputs<name>>) => TypesOf<FunctionOutputs<name>>
      [name in FunctionName]: (...args: any[]) => TypesOf<FunctionOutputs<name>>
    }

    // Note: my IDE actually actually seem to be type-checking these correctly
    // despite the compiler error on the function arg spread above! So close!
    const contract: Contract = {
      // uint256, bool => ()
      addFoobar: (n: number, b: boolean) => [],
      // address => uint256
      getFoobar: (n: Address) => [3],
    }

  })
})
