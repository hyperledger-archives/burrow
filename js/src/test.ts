import {Burrow} from './index';
import * as solc from 'solc';

// convenience function to compile solidity in tests
export const compile = (source: string, name: string) => {
  let desc: solc.InputDescription = {language: 'Solidity', sources: {}};
  desc.sources[name] = {content: source};
  desc.settings = {outputSelection: {'*': {'*': ['*']}}};

  const json = solc.compile(JSON.stringify(desc));
  const compiled: solc.OutputDescription = JSON.parse(json);
  if (compiled.errors) throw new Error(compiled.errors.map(err => err.formattedMessage).toString());
  const contract = compiled.contracts[name][name];
  return {abi: contract.abi, code: {bytecode: contract.evm.bytecode.object, deployedBytecode: contract.evm.deployedBytecode.object} };
}

const url = process.env.BURROW_URL || 'localhost:20997';
const addr = process.env.SIGNING_ADDRESS || 'C9F239591C593CB8EE192B0009C6A0F2C9F8D768';
console.log(`Connecting to Burrow at ${url}...`)
export const burrow = new Burrow(url, addr);

