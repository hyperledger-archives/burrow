import ts, { factory } from 'typescript';
import { BoundsType, CallbackReturnType } from './events';
import {
  AddressType,
  ContractCodecType,
  createCall,
  createCallbackType,
  createParameter,
  createPromiseOf,
  declareConstant,
  ErrorType,
  EventType,
  ExportToken,
  MaybeUint8ArrayType,
  Method,
  StringType,
  Uint8ArrayType,
  UnknownType,
} from './syntax';

export const errName = factory.createIdentifier('err');
export const contractCodecName = factory.createIdentifier('codec');
export const eventName = factory.createIdentifier('event');

export const EventErrParameter = createParameter(errName, ErrorType, undefined, true);
export const EventParameter = createParameter(eventName, EventType, undefined, true);

class Deploy extends Method {
  private abiName = factory.createIdentifier('abi');
  private codeHashName = factory.createIdentifier('codeHash');

  params = [
    createParameter('data', factory.createUnionTypeNode([StringType, Uint8ArrayType])),
    createParameter(
      'contractMeta',
      factory.createArrayTypeNode(
        factory.createTypeLiteralNode([
          factory.createPropertySignature(undefined, this.abiName, undefined, StringType),
          factory.createPropertySignature(undefined, this.codeHashName, undefined, Uint8ArrayType),
        ]),
      ),
      undefined,
      true,
    ),
  ];
  ret = createPromiseOf(AddressType);

  constructor() {
    super('deploy');
  }

  call(
    exp: ts.Expression,
    data: ts.Expression,
    contractMeta?: { abi: ts.Expression; codeHash: ts.Expression }[],
  ): ts.CallExpression {
    return createCall(
      factory.createPropertyAccessExpression(exp, this.id),
      contractMeta
        ? [
            data,
            factory.createArrayLiteralExpression(
              contractMeta.map(({ abi, codeHash }) =>
                factory.createObjectLiteralExpression([
                  factory.createPropertyAssignment(this.abiName, abi),
                  factory.createPropertyAssignment(this.codeHashName, codeHash),
                ]),
              ),
            ),
          ]
        : [data],
    );
  }
}

class Call extends Method {
  params = [
    createParameter('data', factory.createUnionTypeNode([StringType, Uint8ArrayType])),
    createParameter('address', StringType),
  ];
  ret = createPromiseOf(MaybeUint8ArrayType);

  constructor() {
    super('call');
  }

  call(exp: ts.Expression, data: ts.Expression, address: ts.Expression) {
    return createCall(factory.createPropertyAccessExpression(exp, this.id), [data, address]);
  }
}

class CallSim extends Method {
  params = [
    createParameter('data', factory.createUnionTypeNode([StringType, Uint8ArrayType])),
    createParameter('address', StringType),
  ];
  ret = createPromiseOf(MaybeUint8ArrayType);

  constructor() {
    super('callSim');
  }

  call(exp: ts.Expression, data: ts.Expression, address: ts.Expression) {
    return createCall(factory.createPropertyAccessExpression(exp, this.id), [data, address]);
  }
}

class Listen extends Method {
  params = [
    createParameter('signatures', factory.createArrayTypeNode(StringType)),
    createParameter('address', StringType),
    createParameter('callback', createCallbackType([EventErrParameter, EventParameter], CallbackReturnType)),
    createParameter('start', BoundsType, undefined, true),
    createParameter('end', BoundsType, undefined, true),
  ];
  ret = UnknownType;

  constructor() {
    super('listen');
  }

  call(
    exp: ts.Expression,
    sig: ts.Expression,
    addr: ts.Expression,
    callback: ts.Expression,
    start: ts.Expression,
    end: ts.Expression,
  ) {
    return createCall(factory.createPropertyAccessExpression(exp, this.id), [sig, addr, callback, start, end]);
  }
}

class ContractCodec extends Method {
  params = [createParameter('contractABI', StringType)];
  ret = ContractCodecType;

  constructor() {
    super('contractCodec');
  }

  call(provider: ts.Expression, contractABI: ts.Expression) {
    return createCall(factory.createPropertyAccessExpression(provider, this.id), [contractABI]);
  }
}

export class Provider {
  private name = factory.createIdentifier('Provider');

  methods = {
    deploy: new Deploy(),
    call: new Call(),
    callSim: new CallSim(),
    listen: new Listen(),
    contractCodec: new ContractCodec(),
  };

  createInterface(extern?: boolean): ts.InterfaceDeclaration {
    return factory.createInterfaceDeclaration(
      undefined,
      extern ? [ExportToken] : undefined,
      this.name,
      undefined,
      undefined,
      [
        this.methods.deploy.signature(),
        this.methods.call.signature(),
        this.methods.callSim.signature(),
        this.methods.listen.signature(),
        this.methods.contractCodec.signature(),
      ],
    );
  }

  declareContractCodec(client: ts.Identifier, abiName: ts.Identifier): ts.VariableStatement {
    return declareConstant(contractCodecName, this.methods.contractCodec.call(client, abiName));
  }

  type(): ts.TypeReferenceNode {
    return factory.createTypeReferenceNode(this.name);
  }
}

const encodeDeploy = factory.createIdentifier('encodeDeploy');
const encodeFunctionData = factory.createIdentifier('encodeFunctionData');
const decodeFunctionResult = factory.createIdentifier('decodeFunctionResult ');
const decodeEventLog = factory.createIdentifier('decodeEventLog ');

export function callEncodeDeploy(args: ts.Expression[]): ts.CallExpression {
  return createCall(factory.createPropertyAccessExpression(contractCodecName, encodeDeploy), [...args]);
}

export function callEncodeFunctionData(signature: string, args: ts.Expression[]): ts.CallExpression {
  return createCall(factory.createPropertyAccessExpression(contractCodecName, encodeFunctionData), [
    factory.createStringLiteral(signature),
    ...args,
  ]);
}

export function callDecodeFunctionResult(signature: string, data: ts.Expression): ts.CallExpression {
  return createCall(factory.createPropertyAccessExpression(contractCodecName, decodeFunctionResult), [
    factory.createStringLiteral(signature),
    data,
  ]);
}

export function callDecodeEventLog(signature: string, data: ts.Expression, topics: ts.Expression): ts.CallExpression {
  return createCall(factory.createPropertyAccessExpression(contractCodecName, decodeEventLog), [
    factory.createStringLiteral(signature),
    data,
    topics,
  ]);
}
