import { CallTx, ContractMeta, TxInput } from "../../proto/payload_pb";

export const DEFAULT_GAS = 1111111111;

export function callTx(data: string, account: string, address?: string, contractMeta?: ContractMeta[]): CallTx {
  const input = new TxInput();
  input.setAddress(Buffer.from(account, 'hex'));
  input.setAmount(0);

  const payload = new CallTx();
  payload.setInput(input);
  if (address) payload.setAddress(Buffer.from(address, 'hex'));
  payload.setGaslimit(DEFAULT_GAS);
  payload.setFee(0);
  payload.setData(Buffer.from(data, 'hex'));
  // If address is null then we are creating a new contract so provide the contract metadata
  if (!address ) {
    payload.setContractmetaList(contractMeta)
  }

  return payload
}

