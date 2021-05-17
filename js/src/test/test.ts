import { Client } from '../index';

const url = process.env.BURROW_URL || 'localhost:20123';
const addr = process.env.SIGNING_ADDRESS || 'C9F239591C593CB8EE192B0009C6A0F2C9F8D768';
export const burrow = new Client(url, addr);
