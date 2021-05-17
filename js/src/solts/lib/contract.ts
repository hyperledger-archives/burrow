import ts, { ClassDeclaration, factory, MethodDeclaration } from 'typescript';
import { callName } from './caller';
import { decodeName } from './decoder';
import { encodeName } from './encoder';
import { callGetDataFromLog, callGetTopicsFromLog, eventSigHash } from './events';
import { errName, EventErrParameter, LogEventParameter, logName, Provider } from './provider';
import { ContractMethodsList, getRealType, inputOuputsToType, Signature } from './solidity';
import {
  accessThis,
  asRefNode,
  BlockRangeType,
  createCall,
  createCallbackExpression,
  createParameter,
  createPromiseOf,
  declareConstant,
  EventStream,
  ExportToken,
  MaybeUint8ArrayType,
  Method,
  PrivateToken,
  prop,
  PublicToken,
  StringType,
  Undefined,
} from './syntax';

const dataName = factory.createIdentifier('data');
const clientName = factory.createIdentifier('client');
const addressName = factory.createIdentifier('address');
const eventName = factory.createIdentifier('eventName');

export const ContractName = factory.createIdentifier('Contract');

function solidityFunction(name: string, signatures: Signature[], index: number): ts.MethodDeclaration {
  const signature = signatures[index];
  const args = signature.inputs.map((input) => factory.createIdentifier(input.name));
  const encodeFunctionOrOverloadsArray = prop(createCall(encodeName, [accessThis(clientName)]), name);

  // Special case for overloads
  const hasOverloads = signatures.length > 1;

  const encoderFunction = hasOverloads
    ? factory.createElementAccessExpression(encodeFunctionOrOverloadsArray, index)
    : encodeFunctionOrOverloadsArray;

  const decoderFunctionOrOverloadsArray = prop(createCall(decodeName, [accessThis(clientName), dataName]), name);

  const decoderFunction = hasOverloads
    ? factory.createElementAccessExpression(decoderFunctionOrOverloadsArray, index)
    : decoderFunctionOrOverloadsArray;

  const encode = declareConstant(dataName, createCall(encoderFunction, args));

  const returnType = inputOuputsToType(signature.outputs);

  const call = factory.createCallExpression(
    callName,
    [returnType],
    [
      accessThis(clientName),
      accessThis(addressName),
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

  const params = signature.inputs.map((input) => createParameter(input.name, getRealType(input.type)));
  // Suffix overloads
  return new Method(index > 0 ? `${name}_${index}` : name)
    .parameters(params)
    .returns(createPromiseOf(returnType))
    .declaration([encode, factory.createReturnStatement(call)], true);
}

function solidityEvent(name: string, signature: Signature, provider: Provider): ts.MethodDeclaration {
  const callback = factory.createIdentifier('callback');
  const range = factory.createIdentifier('range');
  // Receivers of LogEventParameter
  const data = callGetDataFromLog(logName);
  const topics = callGetTopicsFromLog(logName);
  const decoderFunction = prop(createCall(decodeName, [accessThis(clientName), data, topics]), name);
  return new Method(name)
    .parameter(
      callback,
      createCallbackExpression([
        EventErrParameter,
        createParameter(eventName, inputOuputsToType(signature.inputs), undefined, true),
      ]),
    )
    .parameter(range, BlockRangeType, true)
    .returns(asRefNode(EventStream))
    .declaration([
      factory.createReturnStatement(
        provider.methods.listen.call(
          accessThis(clientName),
          factory.createStringLiteral(eventSigHash(name, signature.inputs)),
          accessThis(addressName),

          factory.createArrowFunction(
            undefined,
            undefined,
            [EventErrParameter, LogEventParameter],
            undefined,
            undefined,
            factory.createBlock([
              factory.createIfStatement(errName, factory.createReturnStatement(createCall(callback, [errName]))),
              factory.createReturnStatement(createCall(callback, [Undefined, createCall(decoderFunction)])),
            ]),
          ),
          range,
        ),
      ),
    ]);
}

function createMethodsFromABI(
  name: string,
  type: 'function' | 'event',
  signatures: Signature[],
  provider: Provider,
): MethodDeclaration[] {
  if (type === 'function') {
    return signatures.map((signature, index) => solidityFunction(name, signatures, index));
  } else if (type === 'event') {
    return [solidityEvent(name, signatures[0], provider)];
  }
  // FIXME: Not sure why this is not inferred since if is exhaustive
  return undefined as never;
}

export function generateContractClass(abi: ContractMethodsList, provider: Provider): ClassDeclaration {
  return factory.createClassDeclaration(undefined, [ExportToken], ContractName, undefined, undefined, [
    factory.createPropertyDeclaration(undefined, [PrivateToken], clientName, undefined, provider.type(), undefined),
    factory.createPropertyDeclaration(undefined, [PublicToken], addressName, undefined, StringType, undefined),
    factory.createConstructorDeclaration(
      undefined,
      undefined,
      [createParameter(clientName, provider.type()), createParameter(addressName, StringType)],
      factory.createBlock(
        [
          factory.createExpressionStatement(factory.createAssignment(accessThis(clientName), clientName)),
          factory.createExpressionStatement(factory.createAssignment(accessThis(addressName), addressName)),
        ],
        true,
      ),
    ),
    ...abi.flatMap((abi) => createMethodsFromABI(abi.name, abi.type, abi.signatures, provider)),
  ]);
}
