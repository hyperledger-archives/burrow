import * as ts from 'typescript';
import { factory } from 'typescript';
import { Provider } from './provider';
import {
  AsyncToken,
  BooleanType,
  ColonToken,
  createCall,
  createParameter,
  declareConstant,
  ExportToken,
  MaybeUint8ArrayType,
  PromiseType,
  QuestionToken,
  StringType,
  Uint8ArrayType,
} from './syntax';

export const defaultCallName = factory.createIdentifier('defaultCall');
export const callerTypeName = factory.createIdentifier('Caller');

export function createCallerFunction(provider: Provider): ts.FunctionDeclaration {
  const output = factory.createIdentifier('Output');
  const client = factory.createIdentifier('client');
  const payload = factory.createIdentifier('payload');
  const returnData = factory.createIdentifier('returnData');
  const data = factory.createIdentifier('data');
  const isSim = factory.createIdentifier('isSim');
  const callback = factory.createIdentifier('callback');
  const addr = factory.createIdentifier('addr');

  return factory.createFunctionDeclaration(
    undefined,
    [ExportToken, AsyncToken],
    undefined,
    defaultCallName,
    [factory.createTypeParameterDeclaration(output)],
    [
      createParameter(client, provider.type()),
      createParameter(addr, StringType),
      createParameter(data, Uint8ArrayType),
      createParameter(isSim, BooleanType),
      createParameter(
        callback,
        factory.createFunctionTypeNode(
          undefined,
          [createParameter(returnData, MaybeUint8ArrayType)],
          factory.createTypeReferenceNode(output, undefined),
        ),
      ),
    ],
    factory.createTypeReferenceNode(PromiseType, [factory.createTypeReferenceNode(output, undefined)]),
    factory.createBlock(
      [
        declareConstant(
          returnData,
          factory.createAwaitExpression(
            factory.createConditionalExpression(
              isSim,
              QuestionToken,
              provider.methods.callSim.call(client, data, addr),
              ColonToken,
              provider.methods.call.call(client, data, addr),
            ),
          ),
        ),
        factory.createReturnStatement(createCall(callback, [returnData])),
      ],
      true,
    ),
  );
}

export const callerTypes: ts.TypeAliasDeclaration[] = [
  factory.createTypeAliasDeclaration(
    undefined,
    [ExportToken],
    callerTypeName,
    undefined,
    factory.createTypeQueryNode(defaultCallName),
  ),
];
