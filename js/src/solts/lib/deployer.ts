import ts, { FunctionDeclaration } from 'typescript';
import { ABI } from './abi';
import { Provider } from './provider';
import { ReplacerName } from './replacer';
import { getRealType, sha3, tokenizeString } from './solidity';
import {
  BufferFrom,
  createCall,
  CreateCallbackDeclaration,
  CreateNewPromise,
  createParameter,
  declareConstant,
  declareLet,
  ExportToken,
  PromiseType,
  rejectOrResolve,
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

export function generateDeployFunction(
  abi: ABI.Func | undefined,
  bin: string,
  links: string[],
  provider: Provider,
): FunctionDeclaration {
  if (bin === '') {
    throw new Error(`Cannot deploy without binary`);
  }

  const parameters = abi ? abi.inputs?.map((input) => createParameter(input.name, getRealType(input.type))) ?? [] : [];
  // const output = ts.createExpressionWithTypeArguments([ts.createTypeReferenceNode(ContractName, [ts.createTypeReferenceNode('Tx', undefined)])], PromiseType);
  const output = ts.createExpressionWithTypeArguments([StringType], PromiseType);

  const statements: ts.Statement[] = [];
  statements.push(declareLet(bytecode, ts.createLiteral(bin)));
  statements.push(
    ...links.map((link) => {
      return ts.createExpressionStatement(
        ts.createAssignment(
          bytecode,
          createCall(ReplacerName, [
            bytecode,
            ts.createStringLiteral('$' + sha3(link).toLowerCase().slice(0, 34) + '$'),
            ts.createIdentifier(tokenizeString(link)),
          ]),
        ),
      );
    }),
  );

  if (abi) {
    const inputs = ts.createArrayLiteral(abi.inputs?.map((arg) => ts.createLiteral(arg.type)));
    const args = abi.inputs?.map((arg) => ts.createIdentifier(arg.name)) ?? [];
    statements.push(
      declareConstant(
        data,
        ts.createBinary(
          bytecode,
          ts.SyntaxKind.PlusToken,
          provider.methods.encode.call(client, ts.createLiteral(''), inputs, ...args),
        ),
      ),
    );
  } else {
    statements.push(declareConstant(data, bytecode));
  }
  statements.push(declareConstant(payload, provider.methods.payload.call(client, data, undefined)));

  const deployFn = provider.methods.deploy.call(
    client,
    payload,
    CreateCallbackDeclaration(
      err,
      addr,
      [
        rejectOrResolve(
          err,
          [
            declareConstant(
              address,
              createCall(
                ts.createPropertyAccess(
                  createCall(ts.createPropertyAccess(BufferFrom(addr), ts.createIdentifier('toString')), [
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
      createParameter(client, ts.createTypeReferenceNode('Provider', [ts.createTypeReferenceNode(type, [])])),
      ...links.map((link) => createParameter(ts.createIdentifier(tokenizeString(link)), StringType)),
      ...parameters,
    ],
    output,
    ts.createBlock(statements, true),
  );
}
