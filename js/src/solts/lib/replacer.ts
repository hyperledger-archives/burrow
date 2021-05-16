import ts from 'typescript';
import { CreateCall, CreateParameter, DeclareConstant, StringType } from './syntax';

export const ReplacerName = ts.createIdentifier('Replace');

export const Replacer = () => {
  const bytecode = ts.createIdentifier('bytecode');
  const name = ts.createIdentifier('name');
  const address = ts.createIdentifier('address');

  const truncated = ts.createIdentifier('truncated');
  const label = ts.createIdentifier('label');

  return ts.createFunctionDeclaration(
    undefined,
    undefined,
    undefined,
    ReplacerName,
    undefined,
    [CreateParameter(bytecode, StringType), CreateParameter(name, StringType), CreateParameter(address, StringType)],
    StringType,
    ts.createBlock(
      [
        ts.createExpressionStatement(
          ts.createAssignment(
            address,
            adds(
              address,
              arrayJoin(
                ts.createAdd(
                  ts.createSubtract(ts.createNumericLiteral('40'), ts.createPropertyAccess(address, 'length')),
                  ts.createNumericLiteral('1'),
                ),
                '0',
              ),
            ),
          ),
        ),
        DeclareConstant(
          truncated,
          CreateCall(ts.createPropertyAccess(name, 'slice'), [
            ts.createNumericLiteral('0'),
            ts.createNumericLiteral('36'),
          ]),
        ),
        DeclareConstant(
          label,
          adds(
            ts.createAdd(ts.createStringLiteral('__'), truncated),
            arrayJoin(
              ts.createSubtract(ts.createNumericLiteral('37'), ts.createPropertyAccess(truncated, 'length')),
              '_',
            ),
            ts.createStringLiteral('__'),
          ),
        ),
        ts.createWhile(
          ts.createBinary(
            CreateCall(ts.createPropertyAccess(bytecode, 'indexOf'), [label]),
            ts.SyntaxKind.GreaterThanEqualsToken,
            ts.createNumericLiteral('0'),
          ),
          ts.createExpressionStatement(
            ts.createAssignment(bytecode, CreateCall(ts.createPropertyAccess(bytecode, 'replace'), [label, address])),
          ),
        ),
        ts.createReturn(bytecode),
      ],
      true,
    ),
  );
};

function adds(...exp: ts.Expression[]) {
  return exp.reduce((all, next) => {
    return ts.createAdd(all, next);
  });
}

function arrayJoin(length: ts.Expression, literal: string) {
  return CreateCall(
    ts.createPropertyAccess(CreateCall(ts.createIdentifier('Array'), [length]), ts.createIdentifier('join')),
    [ts.createStringLiteral(literal)],
  );
}
