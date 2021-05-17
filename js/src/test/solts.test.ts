import * as assert from 'assert';
import { readEvents } from '../events';
import { Addition } from '../solts/sol/Addition.abi';
import { Contract } from '../solts/sol/Event.abi';
import { burrow } from './test';

describe('solts', () => {
  it('can deploy and call from codegen', async () => {
    const address = await Addition.deploy(burrow);
    const add = new Addition.Contract(burrow, address);
    const { sum } = await add.add(2342, 23432);
    assert.strictEqual(sum, 25774);
  });

  it('can receive events', async () => {
    const address = await Contract.deploy(burrow);
    const eventer = new Contract.Contract(burrow, address);
    const height = await burrow.latestHeight();
    await eventer.announce();
    await eventer.announce();
    await eventer.announce();
    const events = await readEvents(eventer.Init.bind(eventer));
    assert.strictEqual(events.length, 3);
    const event = events[0];
    assert.strictEqual(event.controller, 'C9F239591C593CB8EE192B0009C6A0F2C9F8D768');
    assert.strictEqual(event.metadata, 'bacon,beans,eggs,tomato');
  });
});
