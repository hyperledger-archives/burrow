// The convert functions are based on types used by ethers AbiCoder but we redefine some types here to keep our
// functional dependency on those types minimal

// Same as ethers hybrid array/record type used for dynamic returns
type Result<T = any> = readonly T[] & { readonly [key: string]: T };

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

// Converts values from those returned by ABI decoder to those expected by Burrow's GRPC bindings
export function abiToBurrowResult(args: Result, outputs: ParamType[] | undefined): Result {
  const out: Presult = [];
  if (!outputs) {
    return Object.freeze(out);
  }
  checkParamTypesAndArgs('abiToBurrow', outputs, args);
  for (let i = 0; i < outputs.length; i++) {
    pushValue(out, abiToBurrow(args[i], outputs[i].type), outputs[i]);
  }
  return Object.freeze(out);
}

// Converts values from those returned by Burrow's GRPC bindings to those expected by ABI encoder
export function burrowToAbiResult(args: Result, inputs: ParamType[]): Result {
  const out: Presult = [];
  checkParamTypesAndArgs('burrowToAbi', inputs, args);
  for (let i = 0; i < inputs.length; i++) {
    pushValue(out, burrowToAbi(args[i], inputs[i].type), inputs[i]);
  }
  return Object.freeze(out);
}

export function withoutArrayElements(result: Result): Record<string, unknown> {
  return Object.fromEntries(Object.entries(result).filter(([k]) => isNaN(Number(k))));
}

function abiToBurrow(arg: unknown, type: string): unknown {
  if (/address/.test(type)) {
    return recApply(unprefixedHexString, arg as NestedArray<string>);
  } else if (/bytes[0-9]+/.test(type)) {
    // Handle bytes32 differently - for legacy reasons they are used as identifiers and represented as hex strings
    return recApply(unprefixedHexString, arg as NestedArray<string>);
  } else if (/bytes/.test(type)) {
    return recApply(toBuffer, arg as NestedArray<string>);
  } else if (/int/.test(type)) {
    return recApply(numberToBurrow, arg as NestedArray<BigNumber>);
  }
  return arg;
}

function burrowToAbi(arg: unknown, type: string): unknown {
  if (/address/.test(type)) {
    return recApply(prefixedHexString, arg as NestedArray<string>);
  } else if (/bytes/.test(type)) {
    return recApply(toBuffer, arg as NestedArray<string>);
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
export function prefixedHexString(arg: string | Uint8Array): string {
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
