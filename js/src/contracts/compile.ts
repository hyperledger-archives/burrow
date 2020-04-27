import solc from "solc";
import { ABI } from "./abi";

export type CompiledCode = { bytecode: string; deployedBytecode?: string }

interface CompilerOutput {
  abi: ABI;
  code: CompiledCode;
}

// Compile solidity source code
export function compile(source: string, name: string): CompilerOutput {
  const desc: solc.InputDescription = {language: 'Solidity', sources: {}};
  desc.sources[name] = {content: source};
  desc.settings = {outputSelection: {'*': {'*': ['*']}}};

  const json = solc.compile(JSON.stringify(desc));
  const compiled: solc.OutputDescription = JSON.parse(json);
  if (compiled.errors) throw new Error(compiled.errors.map(err => err.formattedMessage).toString());
  const contract = compiled.contracts[name][name];
  return {
    abi: contract.abi,
    code: {
      bytecode: contract.evm.bytecode.object,
      deployedBytecode: contract.evm.deployedBytecode.object,
    }
  };
}
