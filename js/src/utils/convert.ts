import BN from 'bn.js';

export function recApply<A, B>(func: (input: A) => B, args: A | A[]): B | B[] {
  if (Array.isArray(args)) {
    const next = [];
    for (let i = 0; i < args.length; i++) {
      next.push(recApply(func, args[i]));
    }
    return next;
  }
  return func(args);
}

export function addressTB(arg: string): string {
  return arg.toUpperCase();
}

export function addressTA(arg: string): string {
  if (!/^0x/i.test(arg)) {
    return '0x' + arg;
  }
  return arg;
}

export function bytesTB(arg: Buffer): string {
  return arg.toString('hex').toUpperCase();
}

export function bytesTA(arg: string): Buffer {
  if (typeof (arg) === 'string' && /^0x/i.test(arg)) {
    arg = arg.slice(2);
  }
  return Buffer.from(arg, 'hex');
}

export function numberTB(arg: BN): number {
  return arg.toNumber();
}

export function abiToBurrow(puts: string[], args: Array<any>): Array<any> {
  const out = [];
  for (let i = 0; i < puts.length; i++) {
    if (/address/i.test(puts[i])) {
      out.push(recApply(addressTB, args[i]));
    } else if (/bytes/i.test(puts[i])) {
      out.push(recApply(bytesTB, args[i]));
    } else if (/int/i.test(puts[i])) {
      out.push(recApply(numberTB, args[i]));
    } else {
      out.push(args[i]);
    }
  }
  return out
}

export function burrowToAbi(puts: string[], args: string[]): Array<unknown> {
  const out = [];
  for (let i = 0; i < puts.length; i++) {
    if (/address/i.test(puts[i])) {
      out.push(recApply(addressTA, args[i]));
    } else if (/bytes/i.test(puts[i])) {
      out.push(recApply(bytesTA, args[i]));
    } else {
      out.push(args[i]);
    }
  }
  return out;
}
