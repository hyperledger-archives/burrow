export namespace ABI {
  export type Func = {
    type: 'function' | 'constructor' | 'fallback';
    name: string;
    inputs?: Array<FunctionInput>;
    outputs?: Array<FunctionOutput>;
    stateMutability: 'pure' | 'view' | 'nonpayable' | 'payable';
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

  export type FunctionIO = FunctionInput & FunctionOutput;
  export type FunctionOrEvent = Func | Event;
}
