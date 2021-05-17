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
  MaybeUint8ArrayType,
  PromiseType,
  QuestionToken,
  StringType,
  Uint8ArrayType,
} from './syntax';

export const callName = factory.createIdentifier('call');

export function createCallerFunction(provider: Provider): ts.FunctionDeclaration {
  const output = factory.createIdentifier('Output');
  const client = factory.createIdentifier('client');
  const payload = factory.createIdentifier('payload');
  const txe = factory.createIdentifier('txe');
  const data = factory.createIdentifier('data');
  const isSim = factory.createIdentifier('isSim');
  const callback = factory.createIdentifier('callback');
  const addr = factory.createIdentifier('addr');

  return factory.createFunctionDeclaration(
    undefined,
    [AsyncToken],
    undefined,
    callName,
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
          [createParameter('exec', MaybeUint8ArrayType)],
          factory.createTypeReferenceNode(output, undefined),
        ),
      ),
    ],
    factory.createTypeReferenceNode(PromiseType, [factory.createTypeReferenceNode(output, undefined)]),
    factory.createBlock(
      [
        declareConstant(payload, provider.methods.payload.call(client, data, addr)),
        declareConstant(
          txe,
          factory.createAwaitExpression(
            factory.createConditionalExpression(
              isSim,
              QuestionToken,
              provider.methods.callSim.call(client, payload),
              ColonToken,
              provider.methods.call.call(client, payload),
            ),
          ),
        ),
        factory.createReturnStatement(createCall(callback, [txe])),
      ],
      true,
    ),
  );
}
