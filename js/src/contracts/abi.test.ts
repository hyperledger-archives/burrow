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

type FunctionName = abi[number]['name']

type PickFunctionByName<name extends FunctionName> = {
  [i in keyof abi]: abi[i] extends { name: name; type: 'function' } ? abi[i] : never
}[keyof abi]

type Address = string

type Type<T> =
  T extends 'uint256' ? number :
    T extends 'bool' ? boolean :
      T extends 'address' ? Address :
        never

type Arr1 = [
  { type: 'uint256' },
  { type: 'bool' }
]

// **mumble** something about distribution
type PickValue<T, U> = U extends keyof T ? Pick<T, U>[keyof Pick<T, U>] : never;

type TypesOf<T> = { -readonly [k in keyof T]: Type<PickValue<T[k], 'type'>> }

type FunctionInputs<T extends FunctionName> = PickFunctionByName<T>["inputs"]

type FunctionOutputs<T extends FunctionName> = PickFunctionByName<T>["outputs"]

type ProjType<T> = { [k in keyof T]: PickValue<T[k], 'type'> }

type Arr<T> = { [k in keyof T] }

type Mutable<T> =  { -readonly [K in keyof T]: T[K] };

type FunctionType<T extends FunctionName>= (...args: TypesOf<PickFunctionByName<T>['inputs']>) => boolean

type Contract = {
  [name in FunctionName]: (...args: TypesOf<FunctionInputs<name>>) => TypesOf<FunctionOutputs<name>>
}


describe('ABI helpers', () => {
  it('Can extract correct names from ABI', () => {
    const {abi} = compile(source, 'foo')
    console.log(JSON.stringify(abi))
    assert.strictEqual(transformToFullName(abi[0]), "addFoobar(uint256,bool)")
  })
  const b: PickFunctionByName<'getFoobar'> =
    {
      "name": "getFoobar",
      "inputs": [
        {"internalType": "address", "name": "addr", "type": "address"}
      ],
      "outputs": [{"internalType": "uint256", "name": "n", "type": "uint256"}],
      "stateMutability": "view",
      "type": "function"
    }

  const a: FunctionInputs<'addFoobar'> = [
    {"internalType": "uint256", "name": "n", "type": "uint256"},
    {"internalType": "bool", "name": "b", "type": "bool"},
  ]

  const d: ProjType<FunctionInputs<'addFoobar'>> = ["uint256", "bool"]

  const dd: TypesOf<FunctionInputs<'addFoobar'>> = [1, true]
  const ddd: TypesOf<FunctionInputs<'getFoobar'>> = ["adwa"]

  const sss: [number, boolean] = dd

  // const e: []
  it('Function names', () => {

    const aaa: PickValue<Arr1[1], 'type'> = 'bool'

    const mpfoo: TypesOf<Arr1> = [1, true]


    const contract: Contract = {
      addFoobar: (n: number, b: boolean) => [],
      // addFoobar: ["uint256", "bool"],
      getFoobar: (n: Address) => [2],
    }

  })
})
