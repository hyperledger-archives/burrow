import ts from 'typescript';
import { ABI } from './abi';
import { Provider } from './provider';
import { ReplacerName } from './replacer';
import { GetRealType, Hash, TokenizeString } from './solidity';
import {
  BufferFrom,
  CreateCall,
  CreateCallbackDeclaration,
  CreateNewPromise,
  CreateParameter,
  DeclareConstant,
  DeclareLet,
  ExportToken,
  PromiseType,
  RejectOrResolve,
  StringType,
} from './syntax';

const data = ts.createIdentifier('data');
const payload = ts.createIdentifier('payload');
const bytecode = ts.createIdentifier('bytecode');

const err = ts.createIdentifier('err');
const addr = ts.createIdentifier('addr');

const client = ts.createIdentifier('client');
const address = ts.createIdentifier('address');

export const DeployName = ts.createIdentifier('Deploy');

export const Deploy = (abi: ABI.Func, bin: string, links: string[], provider: Provider) => {
  if (bin === '') {
    return undefined;
  }

  const parameters = abi ? abi.inputs.map((input) => CreateParameter(input.name, GetRealType(input.type))) : [];
  // const output = ts.createExpressionWithTypeArguments([ts.createTypeReferenceNode(ContractName, [ts.createTypeReferenceNode('Tx', undefined)])], PromiseType);
  const output = ts.createExpressionWithTypeArguments([StringType], PromiseType);

  const statements: ts.Statement[] = [];
  statements.push(DeclareLet(bytecode, ts.createLiteral(bin)));
  statements.push(
    ...links.map((link) => {
      return ts.createExpressionStatement(
        ts.createAssignment(
          bytecode,
          CreateCall(ReplacerName, [
            bytecode,
            ts.createStringLiteral('$' + Hash(link).toLowerCase().slice(0, 34) + '$'),
            ts.createIdentifier(TokenizeString(link)),
          ]),
        ),
      );
    }),
  );

  if (abi) {
    const inputs = ts.createArrayLiteral(abi.inputs.map((arg) => ts.createLiteral(arg.type)));
    const args = abi.inputs.map((arg) => ts.createIdentifier(arg.name));
    statements.push(
      DeclareConstant(
        data,
        ts.createBinary(
          bytecode,
          ts.SyntaxKind.PlusToken,
          provider.methods.encode.call(client, ts.createLiteral(''), inputs, ...args),
        ),
      ),
    );
  } else {
    statements.push(DeclareConstant(data, bytecode));
  }
  statements.push(DeclareConstant(payload, provider.methods.payload.call(client, data, undefined)));

  const deployFn = provider.methods.deploy.call(
    client,
    payload,
    CreateCallbackDeclaration(
      err,
      addr,
      [
        RejectOrResolve(
          err,
          [
            DeclareConstant(
              address,
              CreateCall(
                ts.createPropertyAccess(
                  CreateCall(ts.createPropertyAccess(BufferFrom(addr), ts.createIdentifier('toString')), [
                    ts.createLiteral('hex'),
                  ]),
                  ts.createIdentifier('toUpperCase'),
                ),
                undefined,
              ),
            ),
          ],
          // [ts.createNew(ContractName, [], [client, address])])
          [address],
        ),
      ],
      undefined,
      true,
    ),
  );

  statements.push(ts.createReturn(CreateNewPromise([ts.createStatement(deployFn)])));

  const type = 'Tx';
  return ts.createFunctionDeclaration(
    undefined,
    [ExportToken],
    undefined,
    DeployName,
    [ts.createTypeParameterDeclaration(type)],
    [
      CreateParameter(client, ts.createTypeReferenceNode('Provider', [ts.createTypeReferenceNode(type, [])])),
      ...links.map((link) => CreateParameter(ts.createIdentifier(TokenizeString(link)), StringType)),
      ...parameters,
    ],
    output,
    ts.createBlock(statements, true),
  );
};
