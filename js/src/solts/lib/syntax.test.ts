import assert from 'assert';
import ts from 'typescript';
import { CreateCallbackExpression, CreateParameter } from './syntax';

function print(node: ts.Node) {
  const printer = ts.createPrinter({ newLine: ts.NewLineKind.LineFeed });
  return printer.printNode(ts.EmitHint.Unspecified, node, undefined);
}

describe('syntax helpers', function () {
  it('should create callback expression', async function () {
    const ErrAndResult = [
      CreateParameter(ts.createIdentifier('err'), undefined),
      CreateParameter(ts.createIdentifier('result'), undefined),
    ];
    assert.equal(print(CreateCallbackExpression(ErrAndResult)), '(err, result) => void');
  });
});
