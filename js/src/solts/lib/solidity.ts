import { Keccak } from 'sha3';
import ts, { TypeNode } from 'typescript';
import { ABI } from './abi';
import { AsArray, AsRefNode, AsTuple, BooleanType, BufferType, NumberType, StringType, VoidType } from './syntax';
import FunctionOrEvent = ABI.FunctionOrEvent;

export function sha3(str: string): string {
  const hash = new Keccak(256).update(str);
  return hash.digest('hex').toUpperCase();
}

export function nameFromABI(abi: ABI.Func | ABI.Event): string {
  if (abi.name.indexOf('(') !== -1) {
    return abi.name;
  }
  const typeName = (abi.inputs as (ABI.EventInput | ABI.FunctionIO)[]).map((i) => i.type).join();
  return abi.name + '(' + typeName + ')';
}

export function getSize(type: string): number {
  return parseInt(type.replace(/.*\[|\].*/gi, ''), 10);
}

export function getRealType(type: string): ts.TypeNode {
  if (/\[\]/i.test(type)) {
    return AsArray(getRealType(type.replace(/\[\]/, '')));
  }
  if (/\[.*\]/i.test(type)) {
    return AsTuple(getRealType(type.replace(/\[.*\]/, '')), getSize(type));
  } else if (/int/i.test(type)) {
    return NumberType;
  } else if (/bool/i.test(type)) {
    return BooleanType;
  } else if (/bytes/i.test(type)) {
    return AsRefNode(BufferType);
  } else {
    return StringType;
  } // address, bytes
}

export function outputToType(sig: Signature): TypeNode {
  if (!sig.outputs?.length) {
    return VoidType;
  }
  const named = sig.outputs.filter((out) => out !== undefined && out.name !== '');
  if (sig.outputs.length === named.length) {
    return ts.createTypeLiteralNode(
      sig.outputs.map((out) =>
        ts.createPropertySignature(undefined, out.name, undefined, getRealType(out.type), undefined),
      ),
    );
  } else {
    return ts.createTupleTypeNode(sig.outputs.map((out) => getRealType(out.type)));
  }
}

export function collapseInputs(signatures: Array<Signature>): Map<string, string[]> {
  return signatures.reduce((args, next) => {
    next.inputs.map((item) => {
      if (!item) {
        return;
      }
      const prev = args.get(item.name);
      args.set(item.name, prev ? [...prev, item.type] : [item.type]);
    });
    return args;
  }, new Map<string, Array<string>>());
}

export function combineTypes(types: Array<string>): TypeNode {
  return types.length === 1 ? getRealType(types[0]) : ts.createUnionTypeNode(types.map((type) => getRealType(type)));
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
        hash: getSigHash(abi),
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
            hash: getSigHash(abi),
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

function getSigHash(abi: FunctionOrEvent): string {
  return sha3(nameFromABI(abi)).slice(0, 8);
}

function getInputs(abi: FunctionOrEvent): InputOutput[] {
  return (
    abi.inputs
      ?.filter((abi) => abi.name !== '')
      .map((abi) => {
        return { name: abi.name, type: abi.type };
      }) ?? []
  );
}
