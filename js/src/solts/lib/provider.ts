import ts, { factory } from 'typescript';
import {
  AddressType,
  asRefNode,
  BlockRangeType,
  CallTxType,
  ContractCodecType,
  createCall,
  createCallbackExpression,
  createParameter,
  createPromiseOf,
  declareConstant,
  EndOfStreamType,
  ErrorType,
  EventStream,
  ExportToken,
  LogEventType,
  MaybeUint8ArrayType,
  Method,
  StringType,
  Uint8ArrayType,
} from './syntax';

export const errName = factory.createIdentifier('err');
export const contractCodecName = factory.createIdentifier('codec');
export const logName = factory.createIdentifier('log');

export const EventErrParameter = createParameter(
  errName,
  factory.createUnionTypeNode([ErrorType, EndOfStreamType]),
  undefined,
  true,
);
export const LogEventParameter = createParameter(logName, LogEventType, undefined, true);

class Deploy extends Method {
  params = [createParameter('msg', CallTxType)];
  ret = createPromiseOf(AddressType);

  constructor() {
    super('deploy');
  }

  call(exp: ts.Expression, tx: ts.Identifier): ts.CallExpression {
    return createCall(factory.createPropertyAccessExpression(exp, this.id), [tx]);
  }
}

class Call extends Method {
  params = [createParameter('msg', CallTxType)];
  ret = createPromiseOf(MaybeUint8ArrayType);

  constructor() {
    super('call');
  }

  call(exp: ts.Expression, tx: ts.Identifier) {
    return createCall(factory.createPropertyAccessExpression(exp, this.id), [tx]);
  }
}

class CallSim extends Method {
  params = [createParameter('msg', CallTxType)];
  ret = createPromiseOf(MaybeUint8ArrayType);

  constructor() {
    super('callSim');
  }

  call(exp: ts.Expression, tx: ts.Identifier) {
    return createCall(factory.createPropertyAccessExpression(exp, this.id), [tx]);
  }
}

class Listen extends Method {
  params = [
    createParameter('signature', StringType),
    createParameter('address', StringType),
    createParameter('callback', createCallbackExpression([EventErrParameter, LogEventParameter])),
    createParameter('range', BlockRangeType, undefined, true),
  ];
  ret = asRefNode(EventStream);

  constructor() {
    super('listen');
  }

  call(exp: ts.Expression, sig: ts.StringLiteral, addr: ts.Expression, callback: ts.Expression, range: ts.Expression) {
    return createCall(factory.createPropertyAccessExpression(exp, this.id), [sig, addr, callback, range]);
  }
}

class Payload extends Method {
  params = [
    createParameter('data', factory.createUnionTypeNode([StringType, Uint8ArrayType])),
    createParameter('address', StringType, undefined, true),
  ];
  ret = CallTxType;

  constructor() {
    super('payload');
  }

  call(provider: ts.Expression, data: ts.Expression, addr?: ts.Expression) {
    return addr
      ? createCall(factory.createPropertyAccessExpression(provider, this.id), [data, addr])
      : createCall(factory.createPropertyAccessExpression(provider, this.id), [data]);
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
    payload: new Payload(),
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
        this.methods.payload.signature(),
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
