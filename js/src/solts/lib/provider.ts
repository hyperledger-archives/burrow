import ts from 'typescript';
import {
  AnyType,
  AsArray,
  AsRefNode,
  CreateCall,
  CreateCallbackExpression,
  CreateParameter,
  ErrorType,
  Method,
  ReadableType,
  StringType,
  Uint8ArrayType,
  VoidType,
} from './syntax';

export const ErrParameter = CreateParameter(ts.createIdentifier('err'), ErrorType);
export const ExecParameter = CreateParameter(ts.createIdentifier('exec'), Uint8ArrayType);
export const AddrParameter = CreateParameter(ts.createIdentifier('addr'), Uint8ArrayType);
export const EventParameter = CreateParameter(ts.createIdentifier('event'), AnyType);

const type = ts.createIdentifier('Tx');
const typeArgument = ts.createTypeReferenceNode(type, undefined);

class Deploy extends Method {
  params = [
    CreateParameter('msg', typeArgument),
    CreateParameter('callback', CreateCallbackExpression([ErrParameter, AddrParameter])),
  ];
  ret = VoidType;

  constructor() {
    super('deploy');
  }
  call(exp: ts.Expression, tx: ts.Identifier, callback: ts.ArrowFunction) {
    return CreateCall(ts.createPropertyAccess(exp, this.id), [tx, callback]);
  }
}

class Call extends Method {
  params = [
    CreateParameter('msg', typeArgument),
    CreateParameter('callback', CreateCallbackExpression([ErrParameter, ExecParameter])),
  ];
  ret = VoidType;

  constructor() {
    super('call');
  }
  call(exp: ts.Expression, tx: ts.Identifier, callback: ts.ArrowFunction) {
    return CreateCall(ts.createPropertyAccess(exp, this.id), [tx, callback]);
  }
}

class CallSim extends Method {
  params = [
    CreateParameter('msg', typeArgument),
    CreateParameter('callback', CreateCallbackExpression([ErrParameter, ExecParameter])),
  ];
  ret = VoidType;

  constructor() {
    super('callSim');
  }
  call(exp: ts.Expression, tx: ts.Identifier, callback: ts.ArrowFunction) {
    return CreateCall(ts.createPropertyAccess(exp, this.id), [tx, callback]);
  }
}

class Listen extends Method {
  params = [
    CreateParameter('signature', StringType),
    CreateParameter('address', StringType),
    CreateParameter('callback', CreateCallbackExpression([ErrParameter, EventParameter])),
  ];
  ret = AsRefNode(ReadableType);

  constructor() {
    super('listen');
  }
  call(exp: ts.Expression, sig: ts.StringLiteral, addr: ts.Expression, callback: ts.Identifier) {
    return CreateCall(ts.createPropertyAccess(exp, this.id), [sig, addr, callback]);
  }
}

class Payload extends Method {
  params = [CreateParameter('data', StringType), CreateParameter('address', StringType, undefined, true)];
  ret = typeArgument;

  constructor() {
    super('payload');
  }
  call(exp: ts.Expression, data: ts.Identifier, addr: ts.Expression) {
    return addr
      ? CreateCall(ts.createPropertyAccess(exp, this.id), [data, addr])
      : CreateCall(ts.createPropertyAccess(exp, this.id), [data]);
  }
}

class Encode extends Method {
  params = [
    CreateParameter('name', StringType),
    CreateParameter('inputs', AsArray(StringType)),
    CreateParameter('args', AsArray(AnyType), undefined, false, true),
  ];
  ret = StringType;

  constructor() {
    super('encode');
  }
  call(exp: ts.Expression, name: ts.StringLiteral, inputs: ts.ArrayLiteralExpression, ...args: ts.Identifier[]) {
    return CreateCall(ts.createPropertyAccess(exp, this.id), [name, inputs, ...args]);
  }
}

class Decode extends Method {
  params = [CreateParameter('data', Uint8ArrayType), CreateParameter('outputs', AsArray(StringType))];
  ret = AnyType;

  constructor() {
    super('decode');
  }
  call(exp: ts.Expression, data: ts.Identifier, outputs: ts.ArrayLiteralExpression) {
    return CreateCall(ts.createPropertyAccess(exp, this.id), [data, outputs]);
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
