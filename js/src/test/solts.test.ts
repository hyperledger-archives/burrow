import * as assert from 'assert';
import { readEvents } from '../events';
import { Addition } from '../solts/sol/Addition.abi';
import { Eventer } from '../solts/sol/Eventer.abi';
import { burrow } from './test';

describe('solts', () => {
  it('can deploy and call from codegen', async () => {
    const adder = await Addition.deployContract(burrow, true);
    const { sum } = await adder.functions.add(2342, 23432);
    assert.strictEqual(sum, 25774);
  });

  it('can receive events', async () => {
    const eventer = await Eventer.deployContract(burrow, true);
    await eventer.functions.announce();
    await eventer.functions.announce();
    await eventer.functions.announce();
    const events = await readEvents(eventer.listeners.Init);
    assert.strictEqual(events.length, 3);
    const event = events[0];
    assert.strictEqual(event.controller, 'C9F239591C593CB8EE192B0009C6A0F2C9F8D768');
    assert.strictEqual(event.metadata, 'bacon,beans,eggs,tomato');
  });

  it('can listen to multiple events', async () => {
    const eventer = await Eventer.deployContract(burrow, true);
    await eventer.functions.announce();
    await eventer.functions.announce();
    const listener = eventer.listenerFor(['MonoRampage', 'Init']);
    const events = await readEvents(listener);
    assert.strictEqual(events.length, 4);
    // Look ma, type narrowing!
    events.map((event) => {
      if (event.name === 'Init') {
        const eventId = Buffer.from('6576656E74310000000000000000000000000000000000000000000000000000', 'hex');
        assert.deepStrictEqual(event.payload.eventId, eventId);
      } else if (event.name === 'MonoRampage') {
        assert.strictEqual(event.payload.timestamp, 123);
      }
    });
  });
});
