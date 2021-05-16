import ts from 'typescript';
import { CallName } from './caller';
import { DecodeName } from './decoder';
import { EncodeName } from './encoder';
import { ErrParameter, EventParameter, Provider } from './provider';
import { CollapseInputs, CombineTypes, ContractMethodsList, OutputToType, Signature } from './solidity';
import {
  AccessThis,
  AsRefNode,
  CreateCall,
  CreateCallbackExpression,
  CreateParameter,
  DeclareConstant,
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

function SolidityFunction(name: string, signatures: Signature[]) {
  const args = Array.from(CollapseInputs(signatures).keys()).map((key) => ts.createIdentifier(key));
  const encode = DeclareConstant(
    data,
    CreateCall(ts.createPropertyAccess(CreateCall(EncodeName, [AccessThis(client)]), name), args),
  );

  const call = ts.createCall(
    CallName,
    [ts.createTypeReferenceNode('Tx', undefined), OutputToType(signatures[0])],
    [
      AccessThis(client),
      AccessThis(address),
      data,
      ts.createLiteral(signatures[0].constant),
      ts.createArrowFunction(
        undefined,
        undefined,
        [CreateParameter(exec, Uint8ArrayType)],
        undefined,
        undefined,
        ts.createBlock(
          [
            ts.createReturn(
              CreateCall(ts.createPropertyAccess(CreateCall(DecodeName, [AccessThis(client), exec]), name), []),
            ),
          ],
          true,
        ),
      ),
    ],
  );

  const params = Array.from(CollapseInputs(signatures), ([key, value]) => CreateParameter(key, CombineTypes(value)));
  return new Method(name).parameters(params).declaration([encode, ts.createReturn(call)], true);
}

function SolidityEvent(name: string, provider: Provider) {
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

function createMethodFromABI(name: string, type: 'function' | 'event', signatures: Signature[], provider: Provider) {
  if (type === 'function') {
    return SolidityFunction(name, signatures);
  } else if (type === 'event') {
    return SolidityEvent(name, provider);
  }
}

export const Contract = (abi: ContractMethodsList, provider: Provider) => {
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
        [CreateParameter(client, provider.getTypeNode()), CreateParameter(address, StringType)],
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
};
