import ts, { ConciseBody, factory, MethodDeclaration, TypeNode } from 'typescript';

export const ErrorType = factory.createTypeReferenceNode('Error');
export const VoidType = factory.createTypeReferenceNode('void', undefined);
export const StringType = factory.createKeywordTypeNode(ts.SyntaxKind.StringKeyword);
export const NumberType = factory.createKeywordTypeNode(ts.SyntaxKind.NumberKeyword);
export const BooleanType = factory.createKeywordTypeNode(ts.SyntaxKind.BooleanKeyword);
export const UnknownType = factory.createKeywordTypeNode(ts.SyntaxKind.UnknownKeyword);
export const UndefinedType = factory.createKeywordTypeNode(ts.SyntaxKind.UndefinedKeyword);
export const Uint8ArrayType = factory.createTypeReferenceNode('Uint8Array');
export const MaybeUint8ArrayType = factory.createUnionTypeNode([Uint8ArrayType, UndefinedType]);

export const PromiseType = factory.createIdentifier('Promise');
export const ReadableType = factory.createIdentifier('Readable');
export const BufferType = factory.createIdentifier('Buffer');
export const Address = factory.createIdentifier('Address');
export const Event = factory.createIdentifier('Event');
export const ContractCodec = factory.createIdentifier('ContractCodec');
export const Result = factory.createIdentifier('Result');
export const CancelStreamSignal = factory.createIdentifier('CancelStreamSignal');
export const ReturnType = factory.createIdentifier('ReturnType');
export const listenerForName = factory.createIdentifier('listenerFor');
export const linkerName = factory.createIdentifier('linker');

export const TType = factory.createTypeReferenceNode('T');
export const AddressType = factory.createTypeReferenceNode(Address);
export const EventType = factory.createTypeReferenceNode(Event);
export const ContractCodecType = factory.createTypeReferenceNode(ContractCodec);
export const CancelStreamSignalType = factory.createTypeReferenceNode(CancelStreamSignal);

export const PrivateToken = factory.createToken(ts.SyntaxKind.PrivateKeyword);
export const PublicToken = factory.createToken(ts.SyntaxKind.PublicKeyword);
export const ExportToken = factory.createToken(ts.SyntaxKind.ExportKeyword);
export const EllipsisToken = factory.createToken(ts.SyntaxKind.DotDotDotToken);
export const QuestionToken = factory.createToken(ts.SyntaxKind.QuestionToken);
export const QuestionDotToken = factory.createToken(ts.SyntaxKind.QuestionDotToken);
export const ColonToken = factory.createToken(ts.SyntaxKind.ColonToken);
export const AsyncToken = factory.createToken(ts.SyntaxKind.AsyncKeyword);
export const ReadonlyToken = factory.createToken(ts.SyntaxKind.ReadonlyKeyword);
export const EqualsGreaterThanToken = factory.createToken(ts.SyntaxKind.EqualsGreaterThanToken);
export const Undefined = factory.createIdentifier('undefined');

export function createCall(fn: ts.Expression | string, args?: ts.Expression[]): ts.CallExpression {
  return factory.createCallExpression(asExp(fn), undefined, args);
}

export function accessThis(name: ts.Identifier): ts.PropertyAccessExpression {
  return factory.createPropertyAccessExpression(factory.createThis(), name);
}

export function bufferFrom(...args: ts.Expression[]): ts.CallExpression {
  return createCall(factory.createPropertyAccessExpression(BufferType, factory.createIdentifier('from')), args);
}

export function asArray(type: ts.TypeNode): ts.ArrayTypeNode {
  return factory.createArrayTypeNode(type);
}

export function asTuple(type: ts.TypeNode, size: number): ts.TupleTypeNode {
  return factory.createTupleTypeNode(Array(size).fill(type));
}

export function asRefNode(id: ts.Identifier): ts.TypeReferenceNode {
  return factory.createTypeReferenceNode(id, undefined);
}

export function asConst(exp: ts.Expression): ts.AsExpression {
  return factory.createAsExpression(exp, factory.createTypeReferenceNode('const'));
}

export function constObject(elements: ts.ObjectLiteralElementLike[]): ts.AsExpression {
  return asConst(factory.createObjectLiteralExpression(elements));
}

export function arrowFunc(params: ts.ParameterDeclaration[], body: ConciseBody): ts.ArrowFunction {
  return factory.createArrowFunction(undefined, undefined, params, undefined, EqualsGreaterThanToken, body);
}

export function hexToBuffer(arg: ts.Expression): ts.CallExpression {
  return createCall(prop(BufferType, 'from'), [arg, factory.createStringLiteral('hex')]);
}

export function arrowFuncT(
  params: ts.ParameterDeclaration[],
  constraint: TypeNode | undefined,
  type: TypeNode | undefined,
  body: ConciseBody,
): ts.ArrowFunction {
  return factory.createArrowFunction(
    undefined,
    [factory.createTypeParameterDeclaration('T', constraint)],
    params,
    type,
    EqualsGreaterThanToken,
    body,
  );
}

export function prop(
  obj: ts.Expression | string,
  name: string | ts.Identifier,
  optionChain?: boolean,
): ts.PropertyAccessExpression {
  return factory.createPropertyAccessChain(asExp(obj), optionChain ? QuestionDotToken : undefined, name);
}

function asExp(exp: ts.Expression | string): ts.Expression {
  return typeof exp === 'string' ? factory.createIdentifier(exp) : exp;
}

export function createParameter(
  name: string | ts.Identifier,
  typeNode?: ts.TypeNode,
  initializer?: ts.Expression,
  isOptional?: boolean,
  isVariadic?: boolean,
): ts.ParameterDeclaration {
  return factory.createParameterDeclaration(
    undefined,
    undefined,
    isVariadic ? EllipsisToken : undefined,
    typeof name === 'string' ? factory.createIdentifier(name) : name,
    isOptional ? QuestionToken : undefined,
    typeNode,
    initializer,
  );
}

export function declareConstant(
  name: ts.Identifier | string,
  initializer?: ts.Expression,
  extern?: boolean,
): ts.VariableStatement {
  return factory.createVariableStatement(
    extern ? [ExportToken] : [],
    factory.createVariableDeclarationList(
      [factory.createVariableDeclaration(name, undefined, undefined, initializer)],
      ts.NodeFlags.Const,
    ),
  );
}

export function declareLet(name: ts.Identifier, initializer: ts.Expression, extern?: boolean): ts.VariableStatement {
  return factory.createVariableStatement(
    extern ? [ExportToken] : [],
    factory.createVariableDeclarationList(
      [factory.createVariableDeclaration(name, undefined, undefined, initializer)],
      ts.NodeFlags.Let,
    ),
  );
}

const resolveFn = factory.createIdentifier('resolve');
const rejectFn = factory.createIdentifier('reject');

export function createPromiseOf(...nodes: ts.TypeNode[]): ts.ExpressionWithTypeArguments {
  return factory.createExpressionWithTypeArguments(PromiseType, nodes);
}

export function createAssignmentStatement(left: ts.Expression, right: ts.Expression): ts.ExpressionStatement {
  return factory.createExpressionStatement(factory.createAssignment(left, right));
}

export function createPromiseBody(error: ts.Identifier, statements: ts.Expression[]): ts.ExpressionStatement {
  return factory.createExpressionStatement(
    factory.createConditionalExpression(
      error,
      QuestionToken,
      createCall(rejectFn, [error]),
      ColonToken,
      createCall(resolveFn, statements ? statements : undefined),
    ),
  );
}

export function rejectOrResolve(error: ts.Identifier, statements: ts.Statement[], success: ts.Expression[]) {
  return factory.createIfStatement(
    error,
    factory.createExpressionStatement(createCall(rejectFn, [error])),
    factory.createBlock([...statements, factory.createExpressionStatement(createCall(resolveFn, success))]),
  );
}

export function CreateNewPromise(
  body: ts.Statement[],
  returnType?: ts.TypeNode,
  multiLine?: boolean,
): ts.NewExpression {
  return factory.createNewExpression(PromiseType, undefined, [
    CreateCallbackDeclaration(resolveFn, rejectFn, body, returnType, multiLine || false),
  ]);
}

export function CreateCallbackDeclaration(
  first: ts.Identifier,
  second: ts.Identifier,
  body: ts.Statement[],
  returnType?: ts.TypeNode,
  multiLine?: boolean,
): ts.ArrowFunction {
  return factory.createArrowFunction(
    undefined,
    undefined,
    [createParameter(first, undefined), createParameter(second, undefined)],
    returnType,
    undefined,
    factory.createBlock(body, multiLine),
  );
}

export function createCallbackType(
  params: ts.ParameterDeclaration[],
  type: ts.TypeNode = VoidType,
): ts.FunctionTypeNode {
  return factory.createFunctionTypeNode(undefined, params, type);
}

function importFrom(pkg: string, ...things: ts.Identifier[]) {
  return factory.createImportDeclaration(
    undefined,
    undefined,
    factory.createImportClause(
      false,
      undefined,
      factory.createNamedImports(things.map((t) => factory.createImportSpecifier(undefined, t))),
    ),
    factory.createStringLiteral(pkg),
  );
}

export function importReadable(): ts.ImportDeclaration {
  return importFrom('stream', ReadableType);
}

export function importBurrow(burrowImportPath: string): ts.ImportDeclaration {
  return importFrom(
    burrowImportPath,
    Address,
    ContractCodec,
    CancelStreamSignal,
    Event,
    listenerForName,
    Result,
    linkerName,
  );
}

export class Method {
  readonly id: ts.Identifier;
  type?: ts.TypeReferenceNode;
  params: ts.ParameterDeclaration[] = [];
  ret?: ts.TypeNode;

  constructor(name: string) {
    this.id = factory.createIdentifier(name);
  }

  parameter(name: string | ts.Identifier, type: ts.TypeNode, optional?: boolean, isVariadic?: boolean): Method {
    this.params.push(createParameter(name, type, undefined, optional, isVariadic));
    return this;
  }

  parameters(arg: ts.ParameterDeclaration | ts.ParameterDeclaration[]): Method {
    if (Array.isArray(arg)) {
      this.params.push(...arg);
    } else {
      this.params.push(arg);
    }
    return this;
  }

  returns(type: ts.TypeNode): Method {
    this.ret = type;
    return this;
  }

  signature(): ts.MethodSignature {
    return factory.createMethodSignature(undefined, this.id, undefined, undefined, this.params, this.ret);
  }

  declaration(statements: ts.Statement[], multiLine?: boolean): MethodDeclaration {
    return factory.createMethodDeclaration(
      undefined,
      undefined,
      undefined,
      this.id,
      undefined,
      undefined,
      this.params,
      this.ret,
      factory.createBlock(statements, multiLine),
    );
  }
}
