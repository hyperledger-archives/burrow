import ts from 'typescript';
import {
  AnyType,
  AsArray,
  AsRefNode,
  createCall,
  CreateCallbackExpression,
  createParameter,
  ErrorType,
  Method,
  ReadableType,
  StringType,
  Uint8ArrayType,
  VoidType,
} from './syntax';

export const ErrParameter = createParameter(ts.createIdentifier('err'), ErrorType);
export const ExecParameter = createParameter(ts.createIdentifier('exec'), Uint8ArrayType);
export const AddrParameter = createParameter(ts.createIdentifier('addr'), Uint8ArrayType);
export const EventParameter = createParameter(ts.createIdentifier('event'), AnyType);

const type = ts.createIdentifier('Tx');
const typeArgument = ts.createTypeReferenceNode(type, undefined);

class Deploy extends Method {
  params = [
    createParameter('msg', typeArgument),
    createParameter('callback', CreateCallbackExpression([ErrParameter, AddrParameter])),
  ];
  ret = VoidType;

  constructor() {
    super('deploy');
  }
  call(exp: ts.Expression, tx: ts.Identifier, callback: ts.ArrowFunction) {
    return createCall(ts.createPropertyAccess(exp, this.id), [tx, callback]);
  }
}

class Call extends Method {
  params = [
    createParameter('msg', typeArgument),
    createParameter('callback', CreateCallbackExpression([ErrParameter, ExecParameter])),
  ];
  ret = VoidType;

  constructor() {
    super('call');
  }
  call(exp: ts.Expression, tx: ts.Identifier, callback: ts.ArrowFunction) {
    return createCall(ts.createPropertyAccess(exp, this.id), [tx, callback]);
  }
}

class CallSim extends Method {
  params = [
    createParameter('msg', typeArgument),
    createParameter('callback', CreateCallbackExpression([ErrParameter, ExecParameter])),
  ];
  ret = VoidType;

  constructor() {
    super('callSim');
  }
  call(exp: ts.Expression, tx: ts.Identifier, callback: ts.ArrowFunction) {
    return createCall(ts.createPropertyAccess(exp, this.id), [tx, callback]);
  }
}

class Listen extends Method {
  params = [
    createParameter('signature', StringType),
    createParameter('address', StringType),
    createParameter('callback', CreateCallbackExpression([ErrParameter, EventParameter])),
  ];
  ret = AsRefNode(ReadableType);

  constructor() {
    super('listen');
  }
  call(exp: ts.Expression, sig: ts.StringLiteral, addr: ts.Expression, callback: ts.Identifier) {
    return createCall(ts.createPropertyAccess(exp, this.id), [sig, addr, callback]);
  }
}

class Payload extends Method {
  params = [createParameter('data', StringType), createParameter('address', StringType, undefined, true)];
  ret = typeArgument;

  constructor() {
    super('payload');
  }
  call(exp: ts.Expression, data: ts.Identifier, addr?: ts.Expression) {
    return addr
      ? createCall(ts.createPropertyAccess(exp, this.id), [data, addr])
      : createCall(ts.createPropertyAccess(exp, this.id), [data]);
  }
}

class Encode extends Method {
  params = [
    createParameter('name', StringType),
    createParameter('inputs', AsArray(StringType)),
    createParameter('args', AsArray(AnyType), undefined, false, true),
  ];
  ret = StringType;

  constructor() {
    super('encode');
  }
  call(exp: ts.Expression, name: ts.StringLiteral, inputs: ts.ArrayLiteralExpression, ...args: ts.Identifier[]) {
    return createCall(ts.createPropertyAccess(exp, this.id), [name, inputs, ...args]);
  }
}

class Decode extends Method {
  params = [createParameter('data', Uint8ArrayType), createParameter('outputs', AsArray(StringType))];
  ret = AnyType;

  constructor() {
    super('decode');
  }
  call(exp: ts.Expression, data: ts.Identifier, outputs: ts.ArrayLiteralExpression) {
    return createCall(ts.createPropertyAccess(exp, this.id), [data, outputs]);
  }
}

export class Provider {
  private name = ts.createIdentifier('Provider');

  methods = {
    deploy: new Deploy(),
    call: new Call(),
    callSim: new CallSim(),
    listen: new Listen(),
    payload: new Payload(),
    encode: new Encode(),
    decode: new Decode(),
  };

  createInterface() {
    return ts.createInterfaceDeclaration(undefined, undefined, this.name, [this.getTypeArgumentDecl()], undefined, [
      this.methods.deploy.signature(),
      this.methods.call.signature(),
      this.methods.callSim.signature(),
      this.methods.listen.signature(),
      this.methods.payload.signature(),
      this.methods.encode.signature(),
      this.methods.decode.signature(),
    ]);
  }

  getType() {
    return type;
  }

  getTypeNode() {
    return ts.createTypeReferenceNode(this.name, [typeArgument]);
  }

  getTypeArgument() {
    return typeArgument;
  }

  getTypeArgumentDecl() {
    return ts.createTypeParameterDeclaration(type);
  }
}
