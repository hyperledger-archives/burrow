import * as fs from 'fs';
import { ResolvedImport } from 'solc';
import { ABI } from './abi';
import InputDescription = Solidity.InputDescription;

export namespace Solidity {
  export type Bytecode = {
    linkReferences: any;
    object: string;
    opcodes: string;
    sourceMap: string;
  };

  export type Contract = {
    assembly: any;
    evm: {
      bytecode: Bytecode;
      deployedBytecode: Bytecode;
    };
    functionHashes: any;
    gasEstimates: any;
    abi: ABI.FunctionOrEvent[];
    opcodes: string;
    runtimeBytecode: string;
    srcmap: string;
    srcmapRuntime: string;
  };

  export type Source = {
    AST: any;
  };

  export type InputDescription = {
    language: string;
    sources: Record<string, { content: string }>;
    settings: {
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

export function encodeInput(obj: Solidity.InputDescription): string {
  return JSON.stringify(obj);
}

export function decodeOutput(str: string): Solidity.OutputDescription {
  return JSON.parse(str);
}

export function inputDescriptionFromFiles(names: string[]): InputDescription {
  const desc = NewInputDescription();
  names.map((name) => {
    desc.sources[name] = { content: fs.readFileSync(name).toString() };
    desc.settings.outputSelection[name] = {};
    desc.settings.outputSelection[name]['*'] = ['*'];
  });
  return desc;
}

export function importLocal(path: string): ResolvedImport {
  return {
    contents: fs.readFileSync(path).toString(),
  };
}

export function tokenizeLinks(links: Record<string, Record<string, unknown>>): string[] {
  const libraries: Array<string> = [];
  for (const file in links) {
    for (const library in links[file]) {
      libraries.push(file + ':' + library);
    }
  }
  return libraries;
}

export function linker(bytecode: string, links: { name: string; address: string }[]): string {
  for (const { name, address } of links) {
    const paddedAddress = address + Array(40 - address.length + 1).join('0');
    const truncated = name.slice(0, 36);
    const label = '__' + truncated + Array(37 - truncated.length).join('_') + '__';
    while (bytecode.indexOf(label) >= 0) {
      bytecode = bytecode.replace(label, paddedAddress);
    }
  }
  return bytecode;
}