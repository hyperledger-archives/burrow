import { Burrow } from './index';
import * as solc from 'solc';

// convenience function to compile solidity in tests
export const compile = (source: string, name: string) => {
  let desc: solc.InputDescription = { language: 'Solidity', sources: {} };
  desc.sources[name] = { content: source };
  desc.settings = { outputSelection: { '*': { '*': ['*'] }}};

  const compiled: solc.OutputDescription = JSON.parse(solc.compile(JSON.stringify(desc)));
  if (compiled.errors) throw new Error(compiled.errors.map(err => err.formattedMessage).toString());
  const contract = compiled.contracts[name][name];
  return { abi: contract.abi, bytecode: contract.evm.bytecode.object };
}

// contract manager test harness
export const Test = () => {
  let burrow: Burrow;

  return {
    before: (callback?: (app: Burrow) => void) => () => {
      const url = process.env.BURROW_URL || 'localhost:10997';
      const addr = process.env.SIGNING_ADDRESS;
      if (!addr) throw new Error("SIGNING_ADDRESS not set.");
      burrow = new Burrow(url, addr);
      if (callback) return callback(burrow);
    },
    it: (callback: (app: Burrow) => void) => () => callback(burrow),
    after: () => () => {},
  }
}