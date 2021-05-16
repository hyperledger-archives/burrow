import ts from 'typescript';
import { Provider } from './provider';
import {
  BooleanType,
  createCall,
  CreateCallbackDeclaration,
  CreateNewPromise,
  createParameter,
  createPromiseBody,
  declareConstant,
  PromiseType,
  StringType,
  Uint8ArrayType,
} from './syntax';

export const CallName = ts.createIdentifier('Call');

export const Caller = (provider: Provider) => {
  const input = provider.getType();
  const output = ts.createIdentifier('Output');
  const client = ts.createIdentifier('client');
  const payload = ts.createIdentifier('payload');
  const data = ts.createIdentifier('data');
  const isSim = ts.createIdentifier('isSim');
  const callback = ts.createIdentifier('callback');
  const err = ts.createIdentifier('err');
  const exec = ts.createIdentifier('exec');
  const addr = ts.createIdentifier('addr');

  return ts.createFunctionDeclaration(
    undefined,
    undefined,
    undefined,
    CallName,
    [ts.createTypeParameterDeclaration(input), ts.createTypeParameterDeclaration(output)],
    [
      createParameter(client, provider.getTypeNode()),
      createParameter(addr, StringType),
      createParameter(data, StringType),
      createParameter(isSim, BooleanType),
      createParameter(
        callback,
        ts.createFunctionTypeNode(
          undefined,
          [createParameter('exec', Uint8ArrayType)],
          ts.createTypeReferenceNode(output, undefined),
        ),
      ),
    ],
    ts.createTypeReferenceNode(PromiseType, [ts.createTypeReferenceNode(output, undefined)]),
    ts.createBlock(
      [
        declareConstant(payload, provider.methods.payload.call(client, data, addr)),
        ts.createIf(
          isSim,
          ts.createReturn(
            CreateNewPromise([
              ts.createExpressionStatement(
                provider.methods.callSim.call(
                  client,
                  payload,
                  CreateCallbackDeclaration(err, exec, [createPromiseBody(err, [createCall(callback, [exec])])]),
                ),
              ),
            ]),
          ),
          ts.createReturn(
            CreateNewPromise([
              ts.createExpressionStatement(
                provider.methods.call.call(
                  client,
                  payload,
                  CreateCallbackDeclaration(err, exec, [createPromiseBody(err, [createCall(callback, [exec])])]),
                ),
              ),
            ]),
          ),
        ),
      ],
      true,
    ),
  );
};
