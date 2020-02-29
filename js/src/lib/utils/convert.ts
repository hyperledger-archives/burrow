import BN from 'bn.js';

export const recApply = function<A, B> (func: (input: A) => B, args: A | A[]): B | B[] {
  if (Array.isArray(args)) {
    let next: any = [];
    for (let i = 0; i < args.length; i++) {
      next.push(recApply(func, args[i]));
    };
    return next;
  }
  return func(args);
}

export const addressTB = function (arg: string) {
  return arg.toUpperCase();
}

export const addressTA = function (arg: string) {
  if (!/^0x/i.test(arg)) {
    return '0x' + arg;
  }
  return arg;
}

export const bytesTB = function (arg: Buffer) {
  return arg.toString('hex').toUpperCase();
}

export const bytesTA = function (arg: string) {
  if (typeof (arg) === 'string' && /^0x/i.test(arg)) {
    arg = arg.slice(2);
  }
  return Buffer.from(arg, 'hex');
}

export const numberTB = function (arg: BN) {
  return arg.toNumber();
}

export const abiToBurrow = function (puts: string[], args: Array<any>) {
  let out: any[] = [];
  for (let i = 0; i < puts.length; i++) {
    if (/address/i.test(puts[i])) {
      out.push(recApply<string, string>(addressTB, args[i]));
    } else if (/bytes/i.test(puts[i])) {
      out.push(recApply<Buffer, string>(bytesTB, args[i]));
    } else if (/int/i.test(puts[i])) {
      out.push(recApply<BN, number>(numberTB, args[i]));
    } else {
      out.push(args[i]);
    }
  }
  return out
}

export const burrowToAbi = function (puts: string[], args: Array<any>) {
  let out = [];
  for (let i = 0; i < puts.length; i++) {
    if (/address/i.test(puts[i])) {
      out.push(recApply<string, string>(addressTA, args[i]));
    } else if (/bytes/i.test(puts[i])) {
      out.push(recApply<string, Buffer>(bytesTA, args[i]));
    } else {
      out.push(args[i]);
    }
  };
  return out;
}