import { Event, EventInput, FunctionInput, FunctionOutput, SolidityFunction } from 'solc';

export type ABI = Array<SolidityFunction | Event>;

export type Address = string;

export type FunctionIO = FunctionInput & FunctionOutput;

// TODO: replace with ethers js

export function transformToFullName(abi: SolidityFunction | Event): string {
  if (abi.name.indexOf('(') !== -1) {
    return abi.name;
  }
  const typeName = (abi.inputs as Array<EventInput | FunctionIO>).map((i) => i.type).join(',');
  return abi.name + '(' + typeName + ')';
}

export function extractDisplayName(name: string): string {
  const length = name.indexOf('(');
  return length !== -1 ? name.substr(0, length) : name;
}

export function extractTypeName(name: string): string {
  /// TODO: make it invulnerable
  const length = name.indexOf('(');
  return length !== -1 ? name.substr(length + 1, name.length - 1 - (length + 1)).replace(' ', '') : '';
}

export function isFunction(abi: SolidityFunction | Event): abi is SolidityFunction {
  return abi.type === 'function' || abi.type === 'constructor';
}
