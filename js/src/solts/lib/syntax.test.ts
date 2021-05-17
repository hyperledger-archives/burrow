import assert from 'assert';
import { factory, Node } from 'typescript';
import { printNodes } from '../api';
import { createCallbackExpression, createParameter, createPromiseOf, PromiseType, StringType } from './syntax';

describe('syntax helpers', function () {
  it('should create callback expression', async function () {
    const ErrAndResult = [
      createParameter(factory.createIdentifier('err'), undefined),
      createParameter(factory.createIdentifier('result'), undefined),
    ];
    assertGenerates(createCallbackExpression(ErrAndResult), '(err, result) => void');
  });

  it('should create promise type', () => {
    assertGenerates(factory.createExpressionWithTypeArguments(PromiseType, [StringType]), 'Promise<string>');
    assertGenerates(createPromiseOf(StringType), 'Promise<string>');
  });
});

function assertGenerates(node: Node, expected: string) {
  assert.strictEqual(printNodes(node), expected);
}
