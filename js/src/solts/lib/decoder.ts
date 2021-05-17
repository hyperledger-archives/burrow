import ts, { factory, VariableStatement } from 'typescript';
import { callDecodeEventLog, callDecodeFunctionResult, Provider } from './provider';
import { ContractMethodsList, inputOuputsToType, InputOutput, Method, Signature } from './solidity';
import { asConst, createParameter, declareConstant, MaybeUint8ArrayType, Uint8ArrayType } from './syntax';

export const decodeName = factory.createIdentifier('decode');
const clientName = factory.createIdentifier('client');
const dataName = factory.createIdentifier('data');
const topicsName = factory.createIdentifier('topics');

export function generateDecodeObject(
  methods: ContractMethodsList,
  provider: Provider,
  abiName: ts.Identifier,
): VariableStatement {
  return generateDecoderObject(methods, provider, abiName, (method) => {
    const decodeFunction = (signature: Signature) => {
      const isFunction = method.type === 'function';
      const inputsOrOutputs = isFunction ? signature.outputs : signature.inputs;
      return factory.createArrowFunction(
        undefined,
        undefined,
        [],
        inputOuputsToType(inputsOrOutputs),
        undefined,
        body(
          isFunction
            ? callDecodeFunctionResult(signature.hash, dataName)
            : callDecodeEventLog(signature.hash, dataName, topicsName),
          inputsOrOutputs,
        ),
      );
    };
    if (method.signatures.length === 1) {
      return decodeFunction(method.signatures[0]);
    }
    return asConst(factory.createArrayLiteralExpression(method.signatures.map(decodeFunction)));
  });
}

function generateDecoderObject(
  methods: ContractMethodsList,
  provider: Provider,
  abiName: ts.Identifier,
  functionMaker: (m: Method) => ts.Expression,
): VariableStatement {
  return declareConstant(
    decodeName,
    factory.createArrowFunction(
      undefined,
      undefined,
      [
        createParameter(clientName, provider.type()),
        createParameter(dataName, MaybeUint8ArrayType),
        createParameter(
          topicsName,
          factory.createArrayTypeNode(Uint8ArrayType),
          factory.createArrayLiteralExpression(),
        ),
      ],
      undefined,
      undefined,
      factory.createBlock([
        provider.declareContractCodec(clientName, abiName),
        factory.createReturnStatement(
          factory.createObjectLiteralExpression(
            methods.map((method) => {
              return factory.createPropertyAssignment(method.name, functionMaker(method));
            }, true),
            true,
          ),
        ),
      ]),
    ),
    true,
  );
}

function body(decodeFn: ts.CallExpression, ios?: InputOutput[]): ts.Block {
  let named = [];
  if (ios && ios[0] !== undefined) {
    named = ios.filter((out) => out.name !== '');
  }

  if (ios?.length && ios.length === named.length) {
    const setter = factory.createVariableStatement(
      [],
      factory.createVariableDeclarationList(
        [
          factory.createVariableDeclaration(
            factory.createArrayBindingPattern(
              ios.map((out) => factory.createBindingElement(undefined, undefined, out.name)),
            ),
            undefined,
            undefined,
            decodeFn,
          ),
        ],
        ts.NodeFlags.Const,
      ),
    );

    return factory.createBlock(
      [
        setter,
        factory.createReturnStatement(
          factory.createObjectLiteralExpression(
            ios.map(({ name }) => factory.createPropertyAssignment(name, factory.createIdentifier(name))),
          ),
        ),
      ],
      true,
    );
  } else {
    return factory.createBlock([factory.createReturnStatement(ios?.length ? decodeFn : undefined)]);
  }
}
