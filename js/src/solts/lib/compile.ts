import * as fs from 'fs';
import { ABI } from './abi';

export namespace Solidity {
  type Bytecode = {
    linkReferences: any;
    object: string;
    opcodes: string;
    sourceMap: string;
  };

  type Contract = {
    assembly: any;
    evm: {
      bytecode: Bytecode;
    };
    functionHashes: any;
    gasEstimates: any;
    abi: ABI.FunctionOrEvent[];
    opcodes: string;
    runtimeBytecode: string;
    srcmap: string;
    srcmapRuntime: string;
  };

  type Source = {
    AST: any;
  };

  export type InputDescription = {
    language: string;
    sources?: Record<string, { content: string }>;
    settings?: {
      outputSelection: Record<string, Record<string, Array<string>>>;
    };
  };

  type Error = {
    sourceLocation?: {
      file: string;
      start: number;
      end: number;
    };
    type: string;
    component: string;
    severity: 'error' | 'warning';
    message: string;
    formattedMessage?: string;
  };

  export type OutputDescription = {
    contracts: Record<string, Record<string, Contract>>;
    errors: Array<Error>;
    sourceList: Array<string>;
    sources: Record<string, Source>;
  };
}

function NewInputDescription(): Solidity.InputDescription {
  return {
    language: 'Solidity',
    sources: {},
    settings: { outputSelection: {} },
  };
}

export const EncodeInput = (obj: Solidity.InputDescription): string => JSON.stringify(obj);
export const DecodeOutput = (str: string): Solidity.OutputDescription => JSON.parse(str);

export function InputDescriptionFromFiles(...names: string[]) {
  const desc = NewInputDescription();
  names.map((name) => {
    desc.sources[name] = { content: fs.readFileSync(name).toString() };
    desc.settings.outputSelection[name] = {};
    desc.settings.outputSelection[name]['*'] = ['*'];
  });
  return desc;
}

export function ImportLocal(path: string) {
  return {
    contents: fs.readFileSync(path).toString(),
  };
}

export function TokenizeLinks(links: Record<string, Record<string, any>>) {
  const libraries: Array<string> = [];
  for (const file in links) {
    for (const library in links[file]) {
      libraries.push(file + ':' + library);
    }
  }
  return libraries;
}
