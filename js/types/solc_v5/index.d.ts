declare module 'solc_v5' {
  export type SolidityFunction = {
    type: 'function' | 'constructor' | 'fallback';
    name: string;
    inputs: Array<FunctionInput>;
    outputs?: Array<FunctionOutput>;
    stateMutability?: 'pure' | 'view' | 'nonpayable' | 'payable';
    payable?: boolean;
    constant?: boolean;
  };

  export type Event = {
    type: 'event';
    name: string;
    inputs: Array<EventInput>;
    anonymous: boolean;
  };

  export type FunctionInput = {
    name: string;
    type: string;
    components?: FunctionInput[];
    internalType?: string;
  };

  export type FunctionOutput = FunctionInput;
  export type EventInput = FunctionInput & { indexed?: boolean };

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
      deployedBytecode: Bytecode;
    };
    functionHashes: any;
    gasEstimates: any;
    abi: (SolidityFunction | Event)[];
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

  export type ResolvedImport = {
    contents: string;
  };

  export type CompilerOptions = {
    import: (path: string) => ResolvedImport;
  };

  export function compile(input: string, opts?: CompilerOptions): string;
}
