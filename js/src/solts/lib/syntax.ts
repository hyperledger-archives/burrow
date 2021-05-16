import ts, { MethodDeclaration, VariableStatement } from 'typescript';

export const Uint8ArrayType = ts.createTypeReferenceNode('Uint8Array', undefined);
export const ErrorType = ts.createTypeReferenceNode('Error', undefined);
export const VoidType = ts.createTypeReferenceNode('void', undefined);
export const StringType = ts.createKeywordTypeNode(ts.SyntaxKind.StringKeyword);
export const NumberType = ts.createKeywordTypeNode(ts.SyntaxKind.NumberKeyword);
export const BooleanType = ts.createKeywordTypeNode(ts.SyntaxKind.BooleanKeyword);
export const AnyType = ts.createKeywordTypeNode(ts.SyntaxKind.AnyKeyword);
export const PromiseType = ts.createIdentifier('Promise');
export const ReadableType = ts.createIdentifier('Readable');
export const BufferType = ts.createIdentifier('Buffer');
export const TupleType = (elements: ts.TypeNode[]) => ts.createTupleTypeNode(elements);

export const PrivateToken = ts.createToken(ts.SyntaxKind.PrivateKeyword);
export const PublicToken = ts.createToken(ts.SyntaxKind.PublicKeyword);
export const ExportToken = ts.createToken(ts.SyntaxKind.ExportKeyword);
export const EllipsisToken = ts.createToken(ts.SyntaxKind.DotDotDotToken);
export const QuestionToken = ts.createToken(ts.SyntaxKind.QuestionToken);

export const createCall = (fn: ts.Expression, args?: ts.Expression[]) => ts.createCall(fn, undefined, args);
export const AccessThis = (name: ts.Identifier) => ts.createPropertyAccess(ts.createThis(), name);
export const BufferFrom = (...args: ts.Expression[]) =>
  createCall(ts.createPropertyAccess(BufferType, ts.createIdentifier('from')), args);
export const AsArray = (type: ts.TypeNode) => ts.createArrayTypeNode(type);
export const AsTuple = (type: ts.TypeNode, size: number) => ts.createTupleTypeNode(Array(size).fill(type));
export const AsRefNode = (id: ts.Identifier) => ts.createTypeReferenceNode(id, undefined);

export function createParameter(
  name: string | ts.Identifier,
  typeNode: ts.TypeNode | undefined,
  initializer?: ts.Expression,
  isOptional?: boolean,
  isVariadic?: boolean,
): ts.ParameterDeclaration {
  return ts.createParameter(
    undefined,
    undefined,
    isVariadic ? EllipsisToken : undefined,
    typeof name === 'string' ? ts.createIdentifier(name) : name,
    isOptional ? QuestionToken : undefined,
    typeNode,
    initializer,
  );
}

export function declareConstant(name: ts.Identifier, initializer?: ts.Expression, extern?: boolean): VariableStatement {
  return ts.createVariableStatement(
    extern ? [ExportToken] : [],
    ts.createVariableDeclarationList([ts.createVariableDeclaration(name, undefined, initializer)], ts.NodeFlags.Const),
  );
}

export function declareLet(name: ts.Identifier, initializer?: ts.Expression, extern?: boolean) {
  return ts.createVariableStatement(
    extern ? [ExportToken] : [],
    ts.createVariableDeclarationList([ts.createVariableDeclaration(name, undefined, initializer)], ts.NodeFlags.Let),
  );
}

const resolveFn = ts.createIdentifier('resolve');
const rejectFn = ts.createIdentifier('reject');

export function createPromiseBody(error: ts.Identifier, statements: ts.Expression[]) {
  return ts.createExpressionStatement(
    ts.createConditional(
      error,
      createCall(rejectFn, [error]),
      createCall(resolveFn, statements ? statements : undefined),
    ),
  );
}

export function rejectOrResolve(error: ts.Identifier, statements: ts.Statement[], success: ts.Expression[]) {
  return ts.createIf(
    error,
    ts.createExpressionStatement(createCall(rejectFn, [error])),
    ts.createBlock([...statements, ts.createExpressionStatement(createCall(resolveFn, success))]),
  );
}

export function CreateNewPromise(
  body: ts.Statement[],
  returnType?: ts.TypeNode,
  multiLine?: boolean,
): ts.NewExpression {
  return ts.createNew(PromiseType, undefined, [
    CreateCallbackDeclaration(resolveFn, rejectFn, body, returnType, multiLine || false),
  ]);
}

export function CreateCallbackDeclaration(
  first: ts.Identifier,
  second: ts.Identifier,
  body: ts.Statement[],
  returnType?: ts.TypeNode,
  multiLine?: boolean,
) {
  return ts.createArrowFunction(
    undefined,
    undefined,
    [createParameter(first, undefined), createParameter(second, undefined)],
    returnType,
    undefined,
    ts.createBlock(body, multiLine),
  );
}

export function CreateCallbackExpression(params: ts.ParameterDeclaration[]) {
  return ts.createFunctionTypeNode(undefined, params, VoidType);
}

function ImportFrom(thing: ts.Identifier, pkg: string) {
  return ts.createImportDeclaration(
    undefined,
    undefined,
    ts.createImportClause(undefined, ts.createNamedImports([ts.createImportSpecifier(undefined, thing)])),
    ts.createLiteral(pkg),
  );
}

export function ImportReadable() {
  return ImportFrom(ReadableType, 'stream');
}

export class Method {
  readonly id: ts.Identifier;
  type?: ts.TypeReferenceNode;
  params: ts.ParameterDeclaration[] = [];
  ret?: ts.TypeNode;

  constructor(name: string) {
    this.id = ts.createIdentifier(name);
  }

  parameter(name: string | ts.Identifier, type: ts.TypeNode, optional?: boolean, isVariadic?: boolean) {
    this.params.push(createParameter(name, type, undefined, optional, isVariadic));
    return this;
  }

  parameters(args: ts.ParameterDeclaration[]) {
    this.params.push(...args);
    return this;
  }

  returns(type: ts.TypeNode) {
    this.ret = type;
    return this;
  }

  signature() {
    return ts.createMethodSignature(undefined, this.params, this.ret, this.id, undefined);
  }

  declaration(statements: ts.Statement[], multiLine?: boolean): MethodDeclaration {
    return ts.createMethod(
      undefined,
      undefined,
      undefined,
      this.id,
      undefined,
      undefined,
      this.params,
      this.ret,
      ts.createBlock(statements, multiLine),
    );
  }
}
