import ts, { Expression, VariableStatement } from 'typescript';
import { Provider } from './provider';
import { ContractMethodsList, outputToType, Signature } from './solidity';
import { createParameter, declareConstant, Uint8ArrayType } from './syntax';

export const DecodeName = ts.createIdentifier('Decode');

function output(decodeFn: ts.CallExpression, sig: Signature): ts.Block {
  let named = [];
  if (sig.outputs && sig.outputs[0] !== undefined) {
    named = sig.outputs.filter((out) => out.name !== '');
  }

  if (sig.outputs?.length && sig.outputs.length === named.length) {
    const setter = ts.createVariableStatement(
      [],
      ts.createVariableDeclarationList(
        [
          ts.createVariableDeclaration(
            ts.createArrayBindingPattern(
              sig.outputs.map((out) => ts.createBindingElement(undefined, undefined, out.name)),
            ),
            undefined,
            decodeFn,
          ),
        ],
        ts.NodeFlags.Const,
      ),
    );

    return ts.createBlock(
      [
        setter,
        ts.createReturn(
          ts.createObjectLiteral(
            sig.outputs.map((out) => ts.createPropertyAssignment(out.name, ts.createIdentifier(out.name))),
          ),
        ),
      ],
      true,
    );
  } else {
    return ts.createBlock([ts.createReturn(sig.outputs?.length ? decodeFn : undefined)]);
  }
}

function decoder(sig: Signature, client: ts.Identifier, provider: Provider, data: ts.Identifier) {
  let args: Expression[] = [];
  if (sig.outputs && sig.outputs[0] !== undefined) {
    args = sig.outputs.map((arg) => ts.createLiteral(arg.type));
  }
  const types = ts.createArrayLiteral(args);
  return provider.methods.decode.call(client, data, types);
}

export function generateDecodeObject(methods: ContractMethodsList, provider: Provider): VariableStatement {
  const client = ts.createIdentifier('client');
  const data = ts.createIdentifier('data');

  return declareConstant(
    DecodeName,
    ts.createArrowFunction(
      undefined,
      [provider.getTypeArgumentDecl()],
      [createParameter(client, provider.getTypeNode()), createParameter(data, Uint8ArrayType)],
      undefined,
      undefined,
      ts.createBlock([
        ts.createReturn(
          ts.createObjectLiteral(
            methods
              .filter((method) => method.type === 'function')
              .map((method) => {
                return ts.createPropertyAssignment(
                  method.name,
                  ts.createArrowFunction(
                    undefined,
                    undefined,
                    [],
                    outputToType(method.signatures[0]),
                    undefined,
                    output(decoder(method.signatures[0], client, provider, data), method.signatures[0]),
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
}
