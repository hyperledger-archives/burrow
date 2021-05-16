import ts, { FunctionDeclaration, VariableStatement } from 'typescript';
import { Provider } from './provider';
import { collapseInputs, combineTypes, ContractMethodsList, Signature } from './solidity';
import { createParameter, declareConstant } from './syntax';

export const EncodeName = ts.createIdentifier('Encode');

function output(signatures: Array<Signature>, client: ts.Identifier, provider: Provider): ts.Block {
  if (signatures.length === 0) {
    return ts.createBlock([ts.createReturn()]);
  } else if (signatures.length === 1) {
    return ts.createBlock([ts.createReturn(encoder(signatures[0], client, provider))]);
  }

  return ts.createBlock(
    signatures
      .filter((sig) => sig.inputs.length > 0)
      .map((sig) => {
        return ts.createIf(
          sig.inputs
            .map((input) =>
              ts.createStrictEquality(ts.createTypeOf(ts.createIdentifier(input.name)), ts.createLiteral('string')),
            )
            .reduce((all, next) => ts.createLogicalAnd(all, next)),
          ts.createReturn(encoder(sig, client, provider)),
        );
      }),
    true,
  );
}

function encoder(sig: Signature, client: ts.Identifier, provider: Provider) {
  const inputs = ts.createArrayLiteral(sig.inputs.map((arg) => ts.createLiteral(arg.type)));
  const args = sig.inputs.map((arg) => ts.createIdentifier(arg.name));
  return provider.methods.encode.call(client, ts.createLiteral(sig.hash), inputs, ...args);
}

export function generateEncodeObject(methods: ContractMethodsList, provider: Provider): VariableStatement {
  const client = ts.createIdentifier('client');

  return declareConstant(
    EncodeName,
    ts.createArrowFunction(
      undefined,
      [provider.getTypeArgumentDecl()],
      [createParameter(client, provider.getTypeNode())],
      undefined,
      undefined,
      ts.createBlock([
        ts.createReturn(
          ts.createObjectLiteral(
            methods
              .filter((method) => method.type === 'function')
              .map(
                (method) =>
                  ts.createPropertyAssignment(
                    method.name,
                    ts.createArrowFunction(
                      undefined,
                      undefined,
                      Array.from(collapseInputs(method.signatures), ([key, value]) =>
                        createParameter(key, combineTypes(value)),
                      ),
                      undefined,
                      undefined,
                      output(method.signatures, client, provider),
                    ),
                  ),
                true,
              ),
            true,
          ),
        ),
      ]),
    ),
    true,
  );
};
