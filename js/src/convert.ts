// The convert functions are based on types used by ethers AbiCoder but we redefine some types here to keep our
// functional dependency on those types minimal

// Same as ethers hybrid array/record type used for dynamic returns
export type Result<T = any> = readonly T[] & { readonly [key: string]: T };

// Bash-in-place version of Result
type Presult<T = any> = T[] & { [key: string]: T };

// Minimal restriction of ethers ParamType
type ParamType = {
  readonly name: string;
  readonly type: string;
};

// Minimal restrictions of ethers BigNumber
type BigNumber = {
  toNumber(): number;
};

const bytesNN = /bytes([0-9]+)/;
const zeroAddress = '0x0000000000000000000000000000000000000000';

// Converts values from those returned by Burrow's GRPC bindings to those expected by ABI encoder
export function preEncodeResult(args: Result, inputs: ParamType[]): Result {
  const out: Presult = [];
  checkParamTypesAndArgs('burrowToAbi', inputs, args);
  for (let i = 0; i < inputs.length; i++) {
    pushValue(out, preEncode(args[i], inputs[i].type), inputs[i]);
  }
  return Object.freeze(out);
}

function preEncode(arg: unknown, type: string): unknown {
  if (/address/.test(type)) {
    return recApply(
      (input) => (input === '0x0' || !input ? zeroAddress : prefixedHexString(input)),
      arg as NestedArray<string>,
    );
  }
  const match = bytesNN.exec(type);
  if (match) {
    // Handle bytes32 differently - for legacy reasons they are used as identifiers and represented as hex strings
    return recApply((input) => {
      return padBytes(input, Number(match[1]));
    }, arg as NestedArray<string>);
  }
  if (/bytes/.test(type)) {
    return recApply(toBuffer, arg as NestedArray<string>);
  }
  return arg;
}

// Converts values from those returned by ABI decoder to those expected by Burrow's GRPC bindings
export function postDecodeResult(args: Result, outputs: ParamType[] | undefined): Result {
  const out: Presult = [];
  if (!outputs) {
    return Object.freeze(out);
  }
  checkParamTypesAndArgs('abiToBurrow', outputs, args);
  for (let i = 0; i < outputs.length; i++) {
    pushValue(out, postDecode(args[i], outputs[i].type), outputs[i]);
  }
  return Object.freeze(out);
}

function postDecode(arg: unknown, type: string): unknown {
  if (/address/.test(type)) {
    return recApply(unprefixedHexString, arg as NestedArray<string>);
  }
  if (bytesNN.test(type)) {
    // Handle bytes32 differently - for legacy reasons they are used as identifiers and represented as hex strings
    return recApply(unprefixedHexString, arg as NestedArray<string>);
  }
  if (/bytes/.test(type)) {
    return recApply(toBuffer, arg as NestedArray<string>);
  }
  if (/int/.test(type)) {
    return recApply(numberToBurrow, arg as NestedArray<BigNumber>);
  }
  return arg;
}

function numberToBurrow(arg: BigNumber): BigNumber | number {
  try {
    // number is limited to 53 bits, BN will throw Error
    return arg.toNumber();
  } catch {
    // arg does not fit into number type, so keep it as BN
    return arg;
  }
}

export function withoutArrayElements(result: Result): Record<string, unknown> {
  return Object.fromEntries(Object.entries(result).filter(([k]) => isNaN(Number(k))));
}

export function unprefixedHexString(arg: string | Uint8Array): string {
  if (arg instanceof Uint8Array) {
    return Buffer.from(arg).toString('hex').toUpperCase();
  }
  if (/^0x/i.test(arg)) {
    return arg.slice(2).toUpperCase();
  }
  return arg.toUpperCase();
}

// Adds a 0x prefix to a hex string if it doesn't have one already or returns a prefixed hex string from bytes
export function prefixedHexString(arg?: string | Uint8Array): string {
  if (!arg) {
    return '';
  }
  if (typeof arg === 'string' && !/^0x/i.test(arg)) {
    return '0x' + arg.toLowerCase();
  }
  if (arg instanceof Uint8Array) {
    return '0x' + Buffer.from(arg).toString('hex').toLowerCase();
  }
  return arg.toLowerCase();
}

// Returns bytes buffer from hex string which may or may not have an 0x prefix
export function toBuffer(arg: string | Uint8Array): Buffer {
  if (arg instanceof Uint8Array) {
    return Buffer.from(arg);
  }
  if (/^0x/i.test(arg)) {
    arg = arg.slice(2);
  }
  return Buffer.from(arg, 'hex');
}

type NestedArray<T> = T | NestedArray<T>[];

// Recursively applies func to an arbitrarily nested array with a single element type as described by solidity tuples
function recApply<A, B>(func: (input: A) => B, args: NestedArray<A>): NestedArray<B> {
  return Array.isArray(args) ? args.map((arg) => recApply(func, arg)) : func(args);
}

function checkParamTypesAndArgs(functionName: string, types: ParamType[], args: ArrayLike<unknown>) {
  if (types.length !== args.length) {
    const quantifier = types.length > args.length ? 'more' : 'fewer';
    const typesString = types.map((t) => t.name).join(', ');
    const argsString = Object.values(args)
      .map((a) => JSON.stringify(a))
      .join(', ');
    throw new Error(
      `${functionName} received ${quantifier} types than arguments: types: [${typesString}], args: [${argsString}]`,
    );
  }
}

function pushValue(out: Presult, value: unknown, type: ParamType): void {
  out.push(value);
  if (type.name) {
    out[type.name] = value;
  }
}

export function padBytes(buf: Uint8Array | string, n: number): Buffer {
  if (typeof buf === 'string') {
    // Parse hex (possible 0x prefixed) into bytes!
    buf = toBuffer(buf);
  }
  if (buf.length > n) {
    throw new Error(`cannot pad buffer ${buf} of length ${buf.length} to ${n} because it is longer than ${n}`);
  }
  const padded = Buffer.alloc(n);
  Buffer.from(buf).copy(padded);
  return padded;
}
