
var recApply = function (arg, func) {
  let newArg
  if (Array.isArray(arg)) {
    newArg = []
    for (var i = 0; i < arg.length; i++) {
      newArg.push(recApply(arg[i], func))
    };
  } else {
    newArg = func(arg)
  }
  return newArg
}

var addressTB = function (arg) {
  return arg.toUpperCase()
}

var addressTA = function (arg) {
  if (!/^0x/i.test(arg)) {
    return '0x' + arg
  }
  return arg
}

var bytesTB = function (arg) {
  return arg.toString('hex').toUpperCase()
}

var bytesTA = function (arg) {
  if (typeof (arg) === 'string' && /^0x/i.test(arg)) {
    arg = arg.slice(2)
  }
  return Buffer.from(arg, 'hex')
  // if (!/^0x/i.test(arg)) {
  //   return '0x' + arg
  // }
  // return arg
}

var numberTB = function (arg) {
  return arg.toNumber()
}

var abiToBurrow = function (puts, args) {
  var out = []
  for (var i = 0; i < puts.length; i++) {
    if (/address/i.test(puts[i])) {
      out.push(recApply(args[i], addressTB))
    } else if (/bytes/i.test(puts[i])) {
      out.push(recApply(args[i], bytesTB))
    } else if (/int/i.test(puts[i])) {
      out.push(recApply(args[i], numberTB))
    } else {
      out.push(args[i])
    }
  };
  return out
}

var burrowToAbi = function (puts, args) {
  var out = []
  for (var i = 0; i < puts.length; i++) {
    if (/address/i.test(puts[i])) {
      out.push(recApply(args[i], addressTA))
    } else if (/bytes/i.test(puts[i])) {
      out.push(recApply(args[i], bytesTA))
    } else {
      out.push(args[i])
    }
  };
  return out
}

module.exports = {
  abiToBurrow: abiToBurrow,
  burrowToAbi: burrowToAbi,
  addressTA: addressTA,
  addressTB: addressTB,
  bytesTA: bytesTA,
  bytesTB: bytesTB,
  numberTB: numberTB,
  recApply: recApply
}
