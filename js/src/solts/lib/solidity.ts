import { Keccak } from 'sha3';
import ts from 'typescript';
import { ABI } from './abi';
import { AsArray, AsRefNode, AsTuple, BooleanType, BufferType, NumberType, StringType, VoidType } from './syntax';

export function Hash(str: string) {
  const hash = new Keccak(256).update(str);
  return hash.digest('hex').toUpperCase();
}

export function NameFromABI(abi: ABI.Func | ABI.Event): string {
  if (abi.name.indexOf('(') !== -1) {
    return abi.name;
  }
  const typeName = (abi.inputs as (ABI.EventInput | ABI.FunctionIO)[]).map((i) => i.type).join();
  return abi.name + '(' + typeName + ')';
}

export function GetSize(type: string) {
  return parseInt(type.replace(/.*\[|\].*/gi, ''), 10);
}

export function GetRealType(type: string): ts.TypeNode {
  if (/\[\]/i.test(type)) {
    return AsArray(GetRealType(type.replace(/\[\]/, '')));
  }
  if (/\[.*\]/i.test(type)) {
    return AsTuple(GetRealType(type.replace(/\[.*\]/, '')), GetSize(type));
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

export function OutputToType(sig: Signature) {
  if (sig.outputs.length === 0) {
    return VoidType;
  }
  const named = sig.outputs.filter((out) => out !== undefined && out.name !== '');
  if (sig.outputs.length === named.length) {
    return ts.createTypeLiteralNode(
      sig.outputs.map((out) =>
        ts.createPropertySignature(undefined, out.name, undefined, GetRealType(out.type), undefined),
      ),
    );
  } else {
    return ts.createTupleTypeNode(sig.outputs.map((out) => GetRealType(out.type)));
  }
}

export function CollapseInputs(signatures: Array<Signature>) {
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

export function CombineTypes(types: Array<string>) {
  return types.length === 1 ? GetRealType(types[0]) : ts.createUnionTypeNode(types.map((type) => GetRealType(type)));
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
  type: MethodType;
  signatures?: Array<Signature>;
};
export type ContractMethods = Map<string, Method>;
export type ContractMethodsList = Array<{ name: string } & Method>;

export function GetContractMethods(abi: ABI.FunctionOrEvent[]) {
  // solidity allows duplicate function names
  return Array.from(
    abi.reduce<ContractMethods>((signatures, abi) => {
      if (abi.name === '') {
        return signatures;
      }
      if (abi.type === 'function') {
        const body = signatures.get(abi.name) || {
          type: 'function',
          signatures: new Array<Signature>(),
        };

        body.signatures.push({
          hash: Hash(NameFromABI(abi)).slice(0, 8),
          inputs: abi.inputs
            .filter((abi) => abi.name !== '')
            .map((abi) => {
              return { name: abi.name, type: abi.type };
            }),
          outputs: abi.outputs.map((abi) => {
            return { name: abi.name, type: abi.type };
          }),
          constant: abi.constant || false,
        });

        signatures.set(abi.name, body);
      } else if (abi.type === 'event') {
        signatures.set(abi.name, { type: 'event' });
      }
      return signatures;
    }, new Map<string, { type: MethodType; constant: boolean }>()),
    ([name, method]) => {
      return { name: name, type: method.type, signatures: method.signatures };
    },
  );
}

export function TokenizeString(input: string) {
  return input.replace(/\W+/g, '_');
}
