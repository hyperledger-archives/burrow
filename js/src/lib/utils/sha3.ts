import { Keccak } from 'sha3';

export default function SHA3(str: string) {
  const hash = (new Keccak(256)).update(str);
  return hash.digest('hex').toUpperCase();
}