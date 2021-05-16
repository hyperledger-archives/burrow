import ts, { ClassDeclaration, MethodDeclaration } from 'typescript';
import { CallName } from './caller';
import { DecodeName } from './decoder';
import { EncodeName } from './encoder';
import { ErrParameter, EventParameter, Provider } from './provider';
import { collapseInputs, combineTypes, ContractMethodsList, outputToType, Signature } from './solidity';
import {
  AccessThis,
  AsRefNode,
  createCall,
  CreateCallbackExpression,
  createParameter,
  declareConstant,
  ExportToken,
  Method,
  PrivateToken,
  PublicToken,
  ReadableType,
  StringType,
  Uint8ArrayType,
} from './syntax';

const exec = ts.createIdentifier('exec');
const data = ts.createIdentifier('data');
const client = ts.createIdentifier('client');
const address = ts.createIdentifier('address');

export const ContractName = ts.createIdentifier('Contract');

function solidityFunction(name: string, signatures: Signature[]): MethodDeclaration {
  const args = Array.from(collapseInputs(signatures).keys()).map((key) => ts.createIdentifier(key));
  const encode = declareConstant(
    data,
    createCall(ts.createPropertyAccess(createCall(EncodeName, [AccessThis(client)]), name), args),
  );

  const call = ts.createCall(
    CallName,
    [ts.createTypeReferenceNode('Tx', undefined), outputToType(signatures[0])],
    [
      AccessThis(client),
      AccessThis(address),
      data,
      ts.createLiteral(signatures[0].constant),
      ts.createArrowFunction(
        undefined,
        undefined,
        [createParameter(exec, Uint8ArrayType)],
        undefined,
        undefined,
        ts.createBlock(
          [
            ts.createReturn(
              createCall(ts.createPropertyAccess(createCall(DecodeName, [AccessThis(client), exec]), name), []),
            ),
          ],
          true,
        ),
      ),
    ],
  );

  const params = Array.from(collapseInputs(signatures), ([key, value]) => createParameter(key, combineTypes(value)));
  return new Method(name).parameters(params).declaration([encode, ts.createReturn(call)], true);
}

function solidityEvent(name: string, provider: Provider): MethodDeclaration {
  const callback = ts.createIdentifier('callback');
  return new Method(name)
    .parameter(callback, CreateCallbackExpression([ErrParameter, EventParameter]))
    .returns(AsRefNode(ReadableType))
    .declaration([
      ts.createReturn(
        provider.methods.listen.call(AccessThis(client), ts.createLiteral(name), AccessThis(address), callback),
      ),
    ]);
}

function createMethodFromABI(
  name: string,
  type: 'function' | 'event',
  signatures: Signature[],
  provider: Provider,
): MethodDeclaration {
  if (type === 'function') {
    return solidityFunction(name, signatures);
  } else if (type === 'event') {
    return solidityEvent(name, provider);
  }
  // FIXME: Not sure why this is not inferred since if is exhaustive
  return undefined as never;
}

export function generateContractClass(abi: ContractMethodsList, provider: Provider): ClassDeclaration {
  return ts.createClassDeclaration(
    undefined,
    [ExportToken],
    ContractName,
    [provider.getTypeArgumentDecl()],
    undefined,
    [
      ts.createProperty(undefined, [PrivateToken], client, undefined, provider.getTypeNode(), undefined),
      ts.createProperty(undefined, [PublicToken], address, undefined, StringType, undefined),
      ts.createConstructor(
        undefined,
        undefined,
        [createParameter(client, provider.getTypeNode()), createParameter(address, StringType)],
        ts.createBlock(
          [
            ts.createStatement(ts.createAssignment(AccessThis(client), client)),
            ts.createStatement(ts.createAssignment(AccessThis(address), address)),
          ],
          true,
        ),
      ),
      ...abi.map((abi) => createMethodFromABI(abi.name, abi.type, abi.signatures, provider)),
    ],
  );
}
