import { Keccak } from 'sha3';
import ts, { factory, TypeNode } from 'typescript';
import { ABI } from './abi';
import { asArray, asRefNode, asTuple, BooleanType, BufferType, NumberType, StringType, VoidType } from './syntax';

export function sha3(str: string): string {
  const hash = new Keccak(256).update(str);
  return hash.digest('hex').toUpperCase();
}

export function nameFromABI(abi: ABI.Func | ABI.Event): string {
  if (abi.name.indexOf('(') !== -1) {
    return abi.name;
  }
  const typeName = (abi.inputs as (ABI.EventInput | ABI.FunctionIO)[]).map((i) => i.type).join(',');
  return abi.name + '(' + typeName + ')';
}

export function getSize(type: string): number {
  return parseInt(type.replace(/.*\[|\].*/gi, ''), 10);
}

export function getRealType(type: string): ts.TypeNode {
  if (/\[\]/i.test(type)) {
    return asArray(getRealType(type.replace(/\[\]/, '')));
  }
  if (/\[.*\]/i.test(type)) {
    return asTuple(getRealType(type.replace(/\[.*\]/, '')), getSize(type));
  } else if (/int/i.test(type)) {
    return NumberType;
  } else if (/bool/i.test(type)) {
    return BooleanType;
  } else if (/bytes/i.test(type)) {
    return asRefNode(BufferType);
  } else {
    return StringType;
  } // address, bytes
}

export function inputOuputsToType(ios?: InputOutput[]): TypeNode {
  if (!ios?.length) {
    return VoidType;
  }
  const named = ios.filter((out) => out !== undefined && out.name !== '');
  if (ios.length === named.length) {
    return factory.createTypeLiteralNode(
      ios.map(({ name, type }) => factory.createPropertySignature(undefined, name, undefined, getRealType(type))),
    );
  } else {
    return factory.createTupleTypeNode(ios.map(({ type }) => getRealType(type)));
  }
}

export type InputOutput = {
  name: string;
  type: string;
};
export type MethodType = 'function' | 'event';
export type Signature = {
  hash: string;
  constant: boolean;
  inputs: Array<InputOutput>;
  outputs?: Array<InputOutput>;
};
export type Method = {
  name: string;
  type: MethodType;
  signatures: Array<Signature>;
};
export type ContractMethods = Map<string, Method>;
export type ContractMethodsList = Array<{ name: string } & Method>;

export function getContractMethods(abi: ABI.FunctionOrEvent[]): Method[] {
  // solidity allows duplicate function names
  const contractMethods = abi.reduce<ContractMethods>((signatures, abi) => {
    if (abi.name === '') {
      return signatures;
    }
    if (abi.type === 'function') {
      const method = signatures.get(abi.name) || {
        name: abi.name,
        type: 'function',
        signatures: [],
      };

      const signature = {
        hash: getFunctionSelector(abi),
        constant: abi.constant || false,
        inputs: getInputs(abi),
        outputs: abi.outputs?.map((abi) => {
          return { name: abi.name, type: abi.type };
        }),
      };
      method.signatures.push(signature);
      signatures.set(method.name, method);
    } else if (abi.type === 'event') {
      signatures.set(abi.name, {
        name: abi.name,
        type: 'event',
        signatures: [
          {
            hash: getEventSignature(abi),
            constant: false,
            inputs: getInputs(abi),
          },
        ],
      });
    }
    return signatures;
  }, new Map<string, Method>());
  return Array.from(contractMethods, ([name, method]) => {
    return { name: name, type: method.type, signatures: method.signatures };
  });
}

export function tokenizeString(input: string): string {
  return input.replace(/\W+/g, '_');
}

function getFunctionSelector(abi: ABI.Func): string {
  return sha3(nameFromABI(abi)).slice(0, 8);
}

function getEventSignature(abi: ABI.Event): string {
  return sha3(nameFromABI(abi));
}

function getInputs(abi: ABI.FunctionOrEvent): InputOutput[] {
  return (
    abi.inputs
      ?.filter((abi) => abi.name !== '')
      .map((abi) => {
        return { name: abi.name, type: abi.type };
      }) ?? []
  );
}
