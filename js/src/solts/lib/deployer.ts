import ts, { factory, FunctionDeclaration, SyntaxKind } from 'typescript';
import { ABI } from '../../contracts/abi';
import { contractFunctionName, contractTypeName } from './contract';
import { callEncodeDeploy, Provider } from './provider';
import { getRealType, sha3, tokenizeString } from './solidity';
import {
  AsyncToken,
  BooleanType,
  BufferType,
  createCall,
  createParameter,
  createPromiseOf,
  declareConstant,
  ExportToken,
  hexToBuffer,
  hexToKeccak256,
  linkerName,
  PromiseType,
  prop,
  StringType,
} from './syntax';

export const deployName = factory.createIdentifier('deploy');
export const deployContractName = factory.createIdentifier('deployContract');
export const bytecodeName = factory.createIdentifier('bytecode');
export const deployedBytecodeName = factory.createIdentifier('deployedBytecode');
export const withContractMetaName = factory.createIdentifier('withContractMeta');

// Variable names
const linkedBytecodeName = factory.createIdentifier('linkedBytecode');
const dataName = factory.createIdentifier('data');
const clientName = factory.createIdentifier('client');

export function generateDeployFunction(
  abi: ABI.Func | undefined,
  links: string[],
  provider: Provider,
  abiName: ts.Identifier,
  contractNames: ts.Identifier[],
): FunctionDeclaration {
  const output = factory.createExpressionWithTypeArguments(PromiseType, [StringType]);

  const statements: ts.Statement[] = [];
  statements.push(provider.declareContractCodec(clientName, abiName));

  let bytecode = bytecodeName;
  const linksName = factory.createIdentifier('links');

  if (links.length) {
    const linksArray = factory.createArrayLiteralExpression(
      links.map((link) =>
        factory.createObjectLiteralExpression([
          factory.createPropertyAssignment(
            'name',
            factory.createStringLiteral('$' + sha3(link).toLowerCase().slice(0, 34) + '$'),
          ),
          factory.createPropertyAssignment(
            'address',
            factory.createIdentifier(tokenizeString(link)),
          ),
        ]),
      ),
    );
    statements.push(declareConstant(linksName, linksArray));
    statements.push(declareConstant(linkedBytecodeName, createCall(linkerName, [bytecodeName, linksName])));

    bytecode = linkedBytecodeName;
  }

  const args = abi?.inputs?.map((arg) => factory.createIdentifier(arg.name)) ?? [];

  statements.push(
    declareConstant(
      dataName,
      createCall(prop(BufferType, 'concat'), [
        factory.createArrayLiteralExpression([hexToBuffer(bytecode), callEncodeDeploy(args)]),
      ]),
    ),
  );

  const contractMeta = contractNames.map((n) => {
    const deployedBytecode = prop(n, deployedBytecodeName);
    return {
      abi: prop(n, abiName),
      codeHash: hexToKeccak256(links.length ? createCall(linkerName, [deployedBytecode, linksName]) : deployedBytecode),
    };
  });

  const deployFn = provider.methods.deploy.call(clientName, dataName, withContractMetaName, contractMeta);

  statements.push(factory.createReturnStatement(deployFn));

  return factory.createFunctionDeclaration(
    undefined,
    [ExportToken],
    undefined,
    deployName,
    undefined,
    deployParameters(abi, links, provider),
    output,
    factory.createBlock(statements, true),
  );
}

export function generateDeployContractFunction(
  abi: ABI.Func | undefined,
  links: string[],
  provider: Provider,
): ts.FunctionDeclaration {
  const parameters = deployParameters(abi, links, provider);
  const addressName = factory.createIdentifier('address');
  const callDeploy = factory.createAwaitExpression(
    createCall(deployName, [
      ...parameters.map((p) => p.name).filter((n): n is ts.Identifier => n.kind === SyntaxKind.Identifier),
    ]),
  );
  return factory.createFunctionDeclaration(
    undefined,
    [ExportToken, AsyncToken],
    undefined,
    deployContractName,
    undefined,
    parameters,
    createPromiseOf(factory.createTypeReferenceNode(contractTypeName)),
    factory.createBlock([
      declareConstant(addressName, callDeploy),
      factory.createReturnStatement(createCall(contractFunctionName, [clientName, addressName])),
    ]),
  );
}

function deployParameters(abi: ABI.Func | undefined, links: string[], provider: Provider): ts.ParameterDeclaration[] {
  const parameters = abi ? abi.inputs?.map((input) => createParameter(input.name, getRealType(input.type))) ?? [] : [];
  return [
    createParameter(clientName, provider.type()),
    ...links.map((link) => createParameter(factory.createIdentifier(tokenizeString(link)), StringType)),
    ...parameters,
    createParameter(withContractMetaName, BooleanType, factory.createFalse()),
  ];
}
