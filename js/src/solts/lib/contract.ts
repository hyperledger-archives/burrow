import ts, { factory, ObjectLiteralElementLike } from 'typescript';
import { defaultCallName } from './caller';
import { decodeName } from './decoder';
import { encodeName } from './encoder';
import {
  BoundsType,
  CallbackReturnType,
  dataFromEvent,
  topicsFromEvent,
  createListener,
  createListenerForFunction,
  eventSigHash,
} from './events';
import { errName, EventErrParameter, eventName, EventParameter, Provider } from './provider';
import { ContractMethodsList, getRealType, inputOuputsToType, Signature } from './solidity';
import {
  asConst,
  constObject,
  createCall,
  createCallbackType,
  createParameter,
  createPromiseOf,
  declareConstant,
  EqualsGreaterThanToken,
  ExportToken,
  MaybeUint8ArrayType,
  Method,
  prop,
  ReturnType,
  StringType,
  Undefined,
  UnknownType,
} from './syntax';

export const contractFunctionName = factory.createIdentifier('contract');
export const contractTypeName = factory.createIdentifier('Contract');
export const functionsGroupName = factory.createIdentifier('functions');
export const listenersGroupName = factory.createIdentifier('listeners');
const dataName = factory.createIdentifier('data');
const clientName = factory.createIdentifier('client');
const addressName = factory.createIdentifier('address');
const listenerForName = factory.createIdentifier('listenerFor');
const listenerName = factory.createIdentifier('listener');

export function declareContractType(): ts.TypeAliasDeclaration {
  return factory.createTypeAliasDeclaration(
    undefined,
    [ExportToken],
    contractTypeName,
    undefined,
    factory.createTypeReferenceNode(ReturnType, [factory.createTypeQueryNode(contractFunctionName)]),
  );
}

export function generateContractObject(
  contractNameName: ts.Identifier,
  abi: ContractMethodsList,
  provider: Provider,
): ts.VariableStatement {
  const functions = abi.filter((a) => a.type === 'function');
  const events = abi.filter((a) => a.type === 'event');

  const functionObjectProperties = functions.length
    ? [
        createGroup(
          functionsGroupName,
          functions.flatMap((a) =>
            a.signatures.map((signature, index) => solidityFunction(a.name, a.signatures, index)),
          ),
        ),
      ]
    : [];

  const eventObjectProperties = events.length
    ? [
        createGroup(
          listenersGroupName,
          events.map((a) => solidityEvent(a.name, a.signatures[0], provider)),
        ),
        factory.createPropertyAssignment(listenerForName, createListenerForFunction(clientName, addressName)),
        factory.createPropertyAssignment(listenerName, createListener(clientName, addressName)),
      ]
    : [];

  return declareConstant(
    contractFunctionName,
    factory.createArrowFunction(
      undefined,
      undefined,
      [createParameter(clientName, provider.type()), createParameter(addressName, StringType)],
      undefined,
      EqualsGreaterThanToken,
      asConst(
        factory.createObjectLiteralExpression([
          factory.createShorthandPropertyAssignment(addressName),
          ...functionObjectProperties,
          ...eventObjectProperties,
        ]),
      ),
    ),
    true,
  );
}

function solidityFunction(name: string, signatures: Signature[], index: number): ts.MethodDeclaration {
  const signature = signatures[index];
  const args = signature.inputs.map((input) => factory.createIdentifier(input.name));
  const encodeFunctionOrOverloadsArray = prop(createCall(encodeName, [clientName]), name);
  const callName = factory.createIdentifier('call');

  // Special case for overloads
  const hasOverloads = signatures.length > 1;

  const encoderFunction = hasOverloads
    ? factory.createElementAccessExpression(encodeFunctionOrOverloadsArray, index)
    : encodeFunctionOrOverloadsArray;

  const decoderFunctionOrOverloadsArray = prop(createCall(decodeName, [clientName, dataName]), name);

  const decoderFunction = hasOverloads
    ? factory.createElementAccessExpression(decoderFunctionOrOverloadsArray, index)
    : decoderFunctionOrOverloadsArray;

  const encode = declareConstant(dataName, createCall(encoderFunction, args));

  const returnType = inputOuputsToType(signature.outputs);

  const call = factory.createCallExpression(
    callName,
    [returnType],
    [
      clientName,
      addressName,
      dataName,
      signature.constant ? factory.createTrue() : factory.createFalse(),
      factory.createArrowFunction(
        undefined,
        undefined,
        [createParameter(dataName, MaybeUint8ArrayType)],
        undefined,
        undefined,
        factory.createBlock([factory.createReturnStatement(createCall(decoderFunction, []))], true),
      ),
    ],
  );

  const callParameter = createParameter(callName, undefined, defaultCallName);

  const params = signature.inputs.map((input) => createParameter(input.name, getRealType(input.type)));
  // Suffix overloads
  return new Method(index > 0 ? `${name}_${index}` : name)
    .parameters(params)
    .parameters(callParameter)
    .returns(createPromiseOf(returnType))
    .declaration([encode, factory.createReturnStatement(call)], true);
}

function solidityEvent(name: string, signature: Signature, provider: Provider): ts.MethodDeclaration {
  const callback = factory.createIdentifier('callback');
  const start = factory.createIdentifier('start');
  const end = factory.createIdentifier('end');
  // Receivers of LogEventParameter
  const data = dataFromEvent(eventName);
  const topics = topicsFromEvent(eventName);
  const decoderFunction = prop(createCall(decodeName, [clientName, data, topics]), name);
  return (
    new Method(name)
      .parameter(
        callback,
        createCallbackType(
          [EventErrParameter, createParameter(eventName, inputOuputsToType(signature.inputs), undefined, true)],
          CallbackReturnType,
        ),
      )
      .parameter(start, BoundsType, true)
      .parameter(end, BoundsType, true)
      // type may be EventStream, allow type assertion without polluting inteface
      .returns(UnknownType)
      .declaration([
        factory.createReturnStatement(
          provider.methods.listen.call(
            clientName,
            factory.createArrayLiteralExpression([factory.createStringLiteral(eventSigHash(name, signature.inputs))]),
            addressName,

            factory.createArrowFunction(
              undefined,
              undefined,
              [EventErrParameter, EventParameter],
              undefined,
              undefined,
              factory.createBlock([
                factory.createIfStatement(errName, factory.createReturnStatement(createCall(callback, [errName]))),
                factory.createReturnStatement(createCall(callback, [Undefined, createCall(decoderFunction)])),
              ]),
            ),
            start,
            end,
          ),
        ),
      ])
  );
}

function createGroup(name: ts.Identifier, elements: ObjectLiteralElementLike[]): ts.PropertyAssignment {
  return factory.createPropertyAssignment(name, constObject(elements));
}
