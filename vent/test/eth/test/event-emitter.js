const EventEmitter = artifacts.require('EventEmitter');

contract('EventEmitter', (accounts) => {
  it('emits events', async () => {
    const eventEmitter = await EventEmitter.deployed();
    const {
      logs: [{ args }],
    } = await eventEmitter.emitTwo();
    assert.equal(args[2], 'Donaudampfschifffahrtselektrizitätenhauptbetriebswerkbauunterbeamtengesellschaft');
  });
});
