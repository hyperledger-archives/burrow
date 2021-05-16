import assert from 'assert';
import ts from 'typescript';
import { printNodes } from '../api';
import { CreateCallbackExpression, createParameter } from './syntax';

describe('syntax helpers', function () {
  it('should create callback expression', async function () {
    const ErrAndResult = [
      createParameter(ts.createIdentifier('err'), undefined),
      createParameter(ts.createIdentifier('result'), undefined),
    ];
    assert.equal(printNodes(CreateCallbackExpression(ErrAndResult)), '(err, result) => void');
  });
});
