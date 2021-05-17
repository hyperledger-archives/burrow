import ts, { factory, VariableStatement } from 'typescript';
import { callEncodeFunctionData, Provider } from './provider';
import { ContractMethodsList, getRealType, Method, Signature } from './solidity';
import { asConst, createParameter, declareConstant } from './syntax';

export const encodeName = factory.createIdentifier('encode');
const client = factory.createIdentifier('client');

export function generateEncodeObject(
  methods: ContractMethodsList,
  provider: Provider,
  abiName: ts.Identifier,
): VariableStatement {
  return generateEncoderObject(encodeName, methods, provider, abiName, (method) => {
    const encodeFunction = (signature: Signature) =>
      factory.createArrowFunction(
        undefined,
        undefined,
        signature.inputs.map((i) => createParameter(i.name, getRealType(i.type))),
        undefined,
        undefined,
        factory.createBlock([
          factory.createReturnStatement(
            callEncodeFunctionData(
              signature.hash,
              signature.inputs.map((arg) => factory.createIdentifier(arg.name)),
            ),
          ),
        ]),
      );
    if (method.signatures.length === 1) {
      return encodeFunction(method.signatures[0]);
    }
    return asConst(factory.createArrayLiteralExpression(method.signatures.map(encodeFunction)));
  });
}

function generateEncoderObject(
  name: ts.Identifier,
  methods: ContractMethodsList,
  provider: Provider,
  abiName: ts.Identifier,
  functionMaker: (m: Method) => ts.Expression,
): VariableStatement {
  return declareConstant(
    name,
    factory.createArrowFunction(
      undefined,
      undefined,
      [createParameter(client, provider.type())],
      undefined,
      undefined,
      factory.createBlock([
        provider.declareContractCodec(client, abiName),
        factory.createReturnStatement(
          factory.createObjectLiteralExpression(
            methods
              .filter((method) => method.type === 'function')
              .map((method) => factory.createPropertyAssignment(method.name, functionMaker(method)), true),
            true,
          ),
        ),
      ]),
    ),
    true,
  );
}
