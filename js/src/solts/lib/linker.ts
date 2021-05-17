import ts, { factory } from 'typescript';
import { createCall, createParameter, declareConstant, StringType } from './syntax';

export const linkerName = factory.createIdentifier('linker');

export function createLinkerFunction(): ts.FunctionDeclaration {
  const bytecode = factory.createIdentifier('bytecode');
  const name = factory.createIdentifier('name');
  const address = factory.createIdentifier('address');

  const truncated = factory.createIdentifier('truncated');
  const label = factory.createIdentifier('label');

  return factory.createFunctionDeclaration(
    undefined,
    undefined,
    undefined,
    linkerName,
    undefined,
    [createParameter(bytecode, StringType), createParameter(name, StringType), createParameter(address, StringType)],
    StringType,
    factory.createBlock(
      [
        factory.createExpressionStatement(
          factory.createAssignment(
            address,
            adds(
              address,
              arrayJoin(
                factory.createAdd(
                  factory.createSubtract(
                    factory.createNumericLiteral('40'),
                    factory.createPropertyAccessExpression(address, 'length'),
                  ),
                  factory.createNumericLiteral('1'),
                ),
                '0',
              ),
            ),
          ),
        ),
        declareConstant(
          truncated,
          createCall(factory.createPropertyAccessExpression(name, 'slice'), [
            factory.createNumericLiteral('0'),
            factory.createNumericLiteral('36'),
          ]),
        ),
        declareConstant(
          label,
          adds(
            factory.createAdd(factory.createStringLiteral('__'), truncated),
            arrayJoin(
              factory.createSubtract(
                factory.createNumericLiteral('37'),
                factory.createPropertyAccessExpression(truncated, 'length'),
              ),
              '_',
            ),
            factory.createStringLiteral('__'),
          ),
        ),
        factory.createWhileStatement(
          factory.createBinaryExpression(
            createCall(factory.createPropertyAccessExpression(bytecode, 'indexOf'), [label]),
            ts.SyntaxKind.GreaterThanEqualsToken,
            factory.createNumericLiteral('0'),
          ),
          factory.createExpressionStatement(
            factory.createAssignment(
              bytecode,
              createCall(factory.createPropertyAccessExpression(bytecode, 'replace'), [label, address]),
            ),
          ),
        ),
        factory.createReturnStatement(bytecode),
      ],
      true,
    ),
  );
}

function adds(...exp: ts.Expression[]) {
  return exp.reduce((all, next) => {
    return factory.createAdd(all, next);
  });
}

function arrayJoin(length: ts.Expression, literal: string) {
  return createCall(
    factory.createPropertyAccessExpression(
      createCall(factory.createIdentifier('Array'), [length]),
      factory.createIdentifier('join'),
    ),
    [factory.createStringLiteral(literal)],
  );
}
