import * as ts from 'typescript';
import { factory } from 'typescript';
import { getRealType, InputOutput, sha3, Signature } from './solidity';
import { createCall, createParameter, prop } from './syntax';

const getDataName = factory.createIdentifier('getData_asU8');
const getTopicsName = factory.createIdentifier('getTopicsList_asU8');

export function generateEventArgs(name: string, signature: Signature): ts.ParameterDeclaration[] {
  return signature.inputs.map(({ name, type }) => createParameter(name, getRealType(type)));
}

export function eventSignature(name: string, inputs: InputOutput[]): string {
  return `${name}(${inputs.map(({ type }) => type).join(',')})`;
}

export function eventSigHash(name: string, inputs: InputOutput[]): string {
  return sha3(eventSignature(name, inputs));
}

export function callGetDataFromLog(log: ts.Expression): ts.CallExpression {
  return createCall(prop(log, getDataName, true));
}

export function callGetTopicsFromLog(log: ts.Expression): ts.CallExpression {
  return createCall(prop(log, getTopicsName, true));
}
