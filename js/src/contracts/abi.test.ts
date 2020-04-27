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

const abiConst = [{
  "inputs": [
    {"internalType": "uint256", "name": "n", "type": "uint256"},
    {"internalType": "bool", "name": "b", "type": "bool"}
  ], "name": "addFoobar", "outputs": [], "stateMutability": "nonpayable", "type": "function"
}, {
  "inputs": [],
  "name": "getFoobar",
  "outputs": [{"internalType": "uint256", "name": "n", "type": "uint256"}],
  "stateMutability": "view",
  "type": "function"
}] as const;

type abi = typeof abiConst
type Func = abi[number]

type PickFunctionByName<T extends string> = Func & { name: T }

type FunctionsByName = {
  [key in abi[number]['name']]: PickFunctionByName<key>
}

type FunctionNames = keyof FunctionsByName

type FunctionInputs = {
  [key in keyof FunctionsByName]: FunctionsByName[key]['inputs']
}

type FunctionInput<Name extends FunctionNames, Index extends number> = FunctionInputs[Name][Index] & { index: Index }

type TypeMapping = {
  'uint256': number;
  'bool': boolean;
}

type InputTypes = FunctionInputs[FunctionNames]


type ff<name extends FunctionNames> = {
  [i in number]: FunctionInput<name, i>['type'] 
}


type Contract = {
  [key in FunctionNames]: FunctionInputs[key]
}

describe('ABI helpers', () => {
  it('Can extract correct names from ABI', () => {
    const {abi} = compile(source, 'foo')
    console.log(JSON.stringify(abi))
    assert.strictEqual(transformToFullName(abi[0]), "addFoobar(uint256,bool)")
  })

  it('Function names', () => {

    const input: ff<'addFoobar'> = ["uint256", "bool", "bool"]
    const input2: FunctionInput<'addFoobar', 0> = {index: 0, "internalType": "uint256", "name": "n", "type": "uint256"}

    const contract: Contract = {
      addFoobar: 'bool',
      getFoobar: undefined,
    }
    const fs: FunctionsByName = {
      addFoobar: {
        "inputs": [
          {"internalType": "uint256", "name": "n", "type": "uint256"},
          {"internalType": "bool", "name": "b", "type": "bool"}
        ],
        "name": "addFoobar",
        "outputs": [],
        "stateMutability": "nonpayable",
        "type": "function"
      },
      getFoobar: {
        "inputs": [],
        "name": "getFoobar",
        "outputs": [{"internalType": "uint256", "name": "n", "type": "uint256"}],
        "stateMutability": "view",
        "type": "function"
      },
    }
    const ins: FunctionInputs = {
      addFoobar: [
        {"internalType": "uint256", "name": "n", "type": "uint256"},
        {"internalType": "bool", "name": "b", "type": "bool"}
      ],
      getFoobar: [],
    }

  })
})
