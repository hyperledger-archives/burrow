import { Keccak } from 'src/utils/sha3';

export default function (str: string): string {
  const hash = (new Keccak(256)).update(str);
  return hash.digest('hex').toUpperCase();
}
