import { Function, FunctionInput, FunctionOutput, Event, EventInput } from 'solc';

type FunctionIO = FunctionInput & FunctionOutput;

export function transformToFullName(abi: Function | Event): string {
  if (abi.name.indexOf('(') !== -1) return abi.name;
  const typeName = (abi.inputs as (EventInput | FunctionIO)[]).map(i => i.type).join();
  return abi.name + '(' + typeName + ')';
}

export function extractDisplayName(name: string): string {
  let length = name.indexOf('(')
  return length !== -1 ? name.substr(0, length) : name
}

export function extractTypeName(name: string): string {
  /// TODO: make it invulnerable
  let length = name.indexOf('(')
  return length !== -1 ? name.substr(length + 1, name.length - 1 - (length + 1)).replace(' ', '') : ''
}

export function isFunction(object: object): boolean {
  return typeof object === 'function'
}