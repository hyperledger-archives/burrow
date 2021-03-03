const EventEmitter = artifacts.require('EventEmitter');

module.exports = function (deployer) {
  deployer.deploy(EventEmitter);
};
