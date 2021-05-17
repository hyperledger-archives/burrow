import ts, { factory, FunctionDeclaration } from 'typescript';
import { ABI } from './abi';
import { linkerName } from './linker';
import { callEncodeDeploy, Provider } from './provider';
import { getRealType, sha3, tokenizeString } from './solidity';
import {
  BufferType,
  createAssignmentStatement,
  createCall,
  createParameter,
  declareConstant,
  declareLet,
  ExportToken,
  PromiseType,
  prop,
  StringType,
} from './syntax';

// Variable names
const payloadName = factory.createIdentifier('payload');
const linkedBytecodeName = factory.createIdentifier('linkedBytecode');
const dataName = factory.createIdentifier('data');
const clientName = factory.createIdentifier('client');

export const DeployName = factory.createIdentifier('deploy');

export function generateDeployFunction(
  abi: ABI.Func | undefined,
  bytecodeName: ts.Identifier,
  links: string[],
  provider: Provider,
  abiName: ts.Identifier,
): FunctionDeclaration {
  const parameters = abi ? abi.inputs?.map((input) => createParameter(input.name, getRealType(input.type))) ?? [] : [];
  const output = factory.createExpressionWithTypeArguments(PromiseType, [StringType]);

  const statements: ts.Statement[] = [];
  statements.push(provider.declareContractCodec(clientName, abiName));
  statements.push(declareLet(linkedBytecodeName, bytecodeName));
  statements.push(
    ...links.map((link) =>
      createAssignmentStatement(
        linkedBytecodeName,
        createCall(linkerName, [
          linkedBytecodeName,
          factory.createStringLiteral('$' + sha3(link).toLowerCase().slice(0, 34) + '$'),
          factory.createIdentifier(tokenizeString(link)),
        ]),
      ),
    ),
  );

  const args = abi?.inputs?.map((arg) => factory.createIdentifier(arg.name)) ?? [];

  statements.push(
    declareConstant(
      dataName,
      createCall(prop(BufferType, 'concat'), [
        factory.createArrayLiteralExpression([
          createCall(prop(BufferType, 'from'), [linkedBytecodeName, factory.createStringLiteral('hex')]),
          callEncodeDeploy(args),
        ]),
      ]),
    ),
  );
  statements.push(declareConstant(payloadName, provider.methods.payload.call(clientName, dataName, undefined)));

  const deployFn = provider.methods.deploy.call(clientName, payloadName);

  statements.push(factory.createReturnStatement(deployFn));

  return factory.createFunctionDeclaration(
    undefined,
    [ExportToken],
    undefined,
    DeployName,
    undefined,
    [
      createParameter(clientName, provider.type()),
      ...links.map((link) => createParameter(factory.createIdentifier(tokenizeString(link)), StringType)),
      ...parameters,
    ],
    output,
    factory.createBlock(statements, true),
  );
}
