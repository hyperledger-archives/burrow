import ts from 'typescript';
import { Provider } from './provider';
import { CollapseInputs, CombineTypes, ContractMethodsList, Signature } from './solidity';
import { CreateParameter, DeclareConstant } from './syntax';

export const EncodeName = ts.createIdentifier('Encode');

function join(...exp: ts.Expression[]) {
  if (exp.length === 0) {
    return undefined;
  }
  return exp.reduce((all, next) => {
    return ts.createLogicalAnd(all, next);
  });
}

function output(signatures: Array<Signature>, client: ts.Identifier, provider: Provider): ts.Block {
  if (signatures.length === 0) {
    return ts.createBlock([ts.createReturn()]);
  } else if (signatures.length === 1) {
    return ts.createBlock([ts.createReturn(encoder(signatures[0], client, provider))]);
  }

  return ts.createBlock(
    [
      ...signatures
        .filter((sig) => sig.inputs.length > 0)
        .map((sig) => {
          return ts.createIf(
            join(
              ...sig.inputs.map((input) => {
                return ts.createStrictEquality(
                  ts.createTypeOf(ts.createIdentifier(input.name)),
                  ts.createLiteral('string'),
                );
              }),
            ),
            ts.createReturn(encoder(sig, client, provider)),
          );
        }),
    ],
    true,
  );
}

function encoder(sig: Signature, client: ts.Identifier, provider: Provider) {
  const inputs = ts.createArrayLiteral(sig.inputs.map((arg) => ts.createLiteral(arg.type)));
  const args = sig.inputs.map((arg) => ts.createIdentifier(arg.name));
  const encodeFn = provider.methods.encode.call(client, ts.createLiteral(sig.hash), inputs, ...args);
  return encodeFn;
}

export const Encode = (methods: ContractMethodsList, provider: Provider) => {
  const client = ts.createIdentifier('client');

  return DeclareConstant(
    EncodeName,
    ts.createArrowFunction(
      undefined,
      [provider.getTypeArgumentDecl()],
      [CreateParameter(client, provider.getTypeNode())],
      undefined,
      undefined,
      ts.createBlock([
        ts.createReturn(
          ts.createObjectLiteral(
            methods
              .filter((method) => method.type === 'function')
              .map((method) => {
                if (method.type !== 'function') {
                  return;
                }
                return ts.createPropertyAssignment(
                  method.name,
                  ts.createArrowFunction(
                    undefined,
                    undefined,
                    Array.from(CollapseInputs(method.signatures), ([key, value]) =>
                      CreateParameter(key, CombineTypes(value)),
                    ),
                    undefined,
                    undefined,
                    output(method.signatures, client, provider),
                  ),
                );
              }, true),
            true,
          ),
        ),
      ]),
    ),
    true,
  );
};
