import { Exception } from '../proto/errors_pb';
import { TxExecution } from '../proto/exec_pb';

export class TxExecutionError extends Error {
  public readonly code: number;

  constructor(exception: Exception) {
    super(exception.getException());
    this.code = exception.getCode();
  }
}

export function getException(txe: TxExecution): TxExecutionError | void {
  const exception = txe.hasException() ? txe.getException() : undefined;
  if (exception) {
    return new TxExecutionError(exception);
  }
}
